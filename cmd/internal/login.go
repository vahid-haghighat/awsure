package internal

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	cfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
	"github.com/google/uuid"
	"gopkg.in/ini.v1"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	AwsSamlEndpoint = "https://signin.aws.amazon.com/saml"
)

func LoginAll() error {
	configs, err := loadConfigs()
	if err != nil {
		return fmt.Errorf("we couldn't find any config files. please run 'awsure config --profile [PROFILE_NAME]' to configure")
	}

	samls := make(map[string]string)
	for _, config := range configs {
		h := config.Hash()
		if _, ok := samls[h]; ok {
			continue
		}

		samls[h], _ = getSaml(config)
	}

	var errs []string
	for profile, _ := range configs {
		err = Login(profile, configs, samls[configs[profile].Hash()])
		if err != nil {
			errs = append(errs, err.Error())
		}
	}

	if errs != nil && len(errs) > 0 {
		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}
	return nil
}

func getSaml(config *configuration) (string, error) {
	loginUrl, err := createLoginUrl(config.AzureAppIdUri, config.AzureTenantId, AwsSamlEndpoint)
	if err != nil {
		return "", err
	}
	saml, err := loginCli(loginUrl, config)
	if err != nil {
		return "", err
	}
	return saml, nil
}

func Login(profile string, configs map[string]*configuration, saml string) error {
	if configs == nil {
		var err error
		configs, err = loadConfigs()
		if err != nil {
			fmt.Println("We couldn't find any config files. Let's take care of that first")
			err = ConfigProfile(profile)
			if err != nil {
				return err
			}
			configs, err = loadConfigs()
			if err != nil {
				return err
			}
		}
	}

	config, foundConfig := configs[profile]
	if !foundConfig {
		return fmt.Errorf("profile %s does not exist", profile)
	}

	jumpRoles, err := loadJumpRoleCredentials()
	if errors.Is(err, fileNotFoundError) {
		jumpRoles = make(map[string]*jumpRoleCredentials)
		jumpRoles[config.DefaultJumpRole] = &jumpRoleCredentials{}
	}
	loggedInJumpRole := jumpRoles[config.DefaultJumpRole]
	now := time.Now()

	if saml == "" {
		saml, err = getSaml(config)
		if err != nil {
			return err
		}
	}
	if loggedInJumpRole == nil || !loggedInJumpRole.AwsExpiration.After(now) {
		var jumpRole *role
		jumpRole, loggedInJumpRole, err = loginToJumpRole(config, saml)
		if err != nil {
			return err
		}
		if jumpRole.roleArn != config.DefaultJumpRole {
			config.DefaultJumpRole = jumpRole.roleArn
			configs[profile] = config
			_ = saveConfig(configs)
		}
		jumpRoles[jumpRole.roleArn] = loggedInJumpRole
		err = saveJumpRoleCredentials(jumpRoles)
		if err != nil {
			return err
		}

	}

	fmt.Printf("Logging in with profile %s\n", profile)

	awsConfig, err := cfg.LoadDefaultConfig(context.Background(), cfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(loggedInJumpRole.AwsAccessKeyId, loggedInJumpRole.AwsSecretAccessKey, loggedInJumpRole.AwsSessionToken)))
	if err != nil {
		return err
	}
	if awsConfig.Region == "" {
		awsConfig.Region = config.Region
	}

	stsClient := sts.NewFromConfig(awsConfig)
	destinationRoleArn := fmt.Sprintf("arn:aws:iam::%s:role/%s", config.DestinationAccountId, config.DestinationRoleName)
	stsInput := sts.AssumeRoleInput{
		RoleArn:         &destinationRoleArn,
		RoleSessionName: &profile,
	}
	awsCredentialsResponse, err := stsClient.AssumeRole(context.Background(), &stsInput)
	if err != nil {
		return err
	}

	awsCredentials, err := ini.Load(defaultAwsCredentialsFileLocation)
	if err != nil {
		fmt.Println("Couldn't find the aws credentials file. Creating a new one")
		err = os.MkdirAll(filepath.Dir(defaultAwsCredentialsFileLocation), 0755)
		if err != nil {
			return nil
		}
		awsCredentials = ini.Empty()
	}
	section := awsCredentials.Section(profile)
	section.Key("aws_access_key_id").SetValue(*awsCredentialsResponse.Credentials.AccessKeyId)
	section.Key("aws_secret_access_key").SetValue(*awsCredentialsResponse.Credentials.SecretAccessKey)
	section.Key("aws_session_token").SetValue(*awsCredentialsResponse.Credentials.SessionToken)
	section.Key("region").SetValue(config.Region)
	section.Key("output").SetValue("json")
	section.Key("aws_expiration").SetValue(awsCredentialsResponse.Credentials.Expiration.Format(timeFormat))

	if err = awsCredentials.SaveTo(defaultAwsCredentialsFileLocation); err != nil {
		return err
	}

	fmt.Printf("Credentials expire at: %s\n\n", awsCredentialsResponse.Credentials.Expiration.Local())
	return nil
}

func getJumpRole(roles []role, config *configuration, err error) (role, error) {
	var rl role

	if len(roles) == 0 {
		return role{}, fmt.Errorf("you don't have access to any role. please contact your administrator to add you to appropriate groups on Azure")
	}

	if config.DefaultJumpRole != "" {
		for _, r := range roles {
			if r.roleArn == config.DefaultJumpRole {
				return r, nil
			}
		}
		fmt.Printf("default jump role %s not found", config.DefaultJumpRole)
	}

	if len(roles) == 1 {
		fmt.Printf("you are assigned to one group. slecting %s\n", roles[0].roleArn)
		rl = roles[0]
	} else {
		prompter := Prompter{}
		if (role{} == rl) {
			var rolesToSelect []string

			sort.SliceStable(roles, func(i, j int) bool {
				return roles[i].roleArn < roles[j].roleArn
			})

			linePrefix := "#"
			for i, r := range roles {
				rolesToSelect = append(rolesToSelect, linePrefix+strconv.Itoa(i+1)+" "+r.roleArn)
			}
			label := fmt.Sprintf("Select your jump role for %s in %s - Hint: fuzzy search supported. To choose one role directly just enter #{Int}", config.DestinationRoleName, config.DestinationAccountId)

			var indexChoice int
			indexChoice, _, err = prompter.Select(label, rolesToSelect, fuzzySearchWithPrefixAnchor(rolesToSelect, linePrefix))
			if err != nil {
				return role{}, err
			}

			rl = roles[indexChoice]
		}
	}
	return rl, nil
}

func loginCli(urlString string, conf *configuration) (string, error) {
	browser := rod.New()

	browser = browser.MustConnect()
	defer browser.MustClose()

	router := browser.HijackRequests()
	defer router.MustStop()

	samlResponseChan := make(chan string, 1)
	samlResult := ""

	router.MustAdd("https://*amazon*", func(ctx *rod.Hijack) {
		reqURL := ctx.Request.URL().String()

		if reqURL == AwsSamlEndpoint {
			val, err := url.ParseQuery(ctx.Request.Body())
			if err != nil {
				fmt.Printf("Fail to saml endpoint response: %v", err)
				os.Exit(1)
			}

			samlResponseChan <- val.Get("SAMLResponse")

			ctx.Response.Fail(proto.NetworkErrorReasonInternetDisconnected)
		} else {
			ctx.ContinueRequest(&proto.FetchContinueRequest{})
		}
	})

	go router.Run()

	stopChan := make(chan struct{})
	go spinner(stopChan)

	page := browser.MustPage()
	wait := page.WaitNavigation(proto.PageLifecycleEventNameDOMContentLoaded)
	page.MustNavigate(urlString)
	wait()

Loop:
	for {
		for _, st := range states {
			select {
			case x, ok := <-samlResponseChan:
				if ok {
					samlResult = x
					stopChan <- struct{}{}
					break Loop
				}
			default:
			}

			el, err := page.Sleeper(rod.NotFoundSleeper).Element(st.selector)

			if err == nil {
				stopChan <- struct{}{}

				err = st.handler(page, el, conf)
				if err != nil {
					return "", err
				}
				stopChan = make(chan struct{})
				go spinner(stopChan)
				time.Sleep(time.Millisecond * 500)
			}
		}
	}

	return samlResult, nil
}

func loginGui(urlString string, conf *configuration) (string, error) {
	stopChan := make(chan struct{})
	go spinner(stopChan)

	l := launcher.New().
		Headless(false).
		Devtools(false)

	defer l.Cleanup()
	controlUrl := l.MustLaunch()

	browser := rod.New()

	browser = browser.ControlURL(controlUrl)

	browser = browser.MustConnect()
	defer browser.MustClose()

	router := browser.HijackRequests()
	defer router.MustStop()

	samlResponseChan := make(chan string, 1)
	samlResult := ""

	router.MustAdd("https://*amazon*", func(ctx *rod.Hijack) {
		reqURL := ctx.Request.URL().String()

		if reqURL == AwsSamlEndpoint {
			val, err := url.ParseQuery(ctx.Request.Body())
			if err != nil {
				fmt.Printf("Fail to saml endpoint response: %v", err)
				os.Exit(1)
			}

			samlResponseChan <- val.Get("SAMLResponse")

			err = browser.Close()
			if err != nil {
				return
			}
		} else {
			ctx.ContinueRequest(&proto.FetchContinueRequest{})
		}
	})

	go router.Run()

	page := browser.MustPage()
	wait := page.WaitNavigation(proto.PageLifecycleEventNameDOMContentLoaded)
	page.MustNavigate(urlString)
	wait()

Loop:
	for {
		select {
		case x, ok := <-samlResponseChan:
			if ok {
				samlResult = x
				stopChan <- struct{}{}
				break Loop
			}
		default:
		}
		time.Sleep(100 * time.Millisecond)
	}

	time.Sleep(1 * time.Second)
	return samlResult, nil
}

func createLoginUrl(appIdUri string, tenantId string, assertionConsumerServiceURL string) (string, error) {
	id := uuid.NewString()

	samlRequest := fmt.Sprintf(`
	<samlp:AuthnRequest xmlns="urn:oasis:names:tc:SAML:2.0:metadata" ID="id%s" Version="2.0" IssueInstant="%s" IsPassive="false" AssertionConsumerServiceURL="%s" xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol">
		<Issuer xmlns="urn:oasis:names:tc:SAML:2.0:assertion">%s</Issuer>
		<samlp:NameIDPolicy Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress"></samlp:NameIDPolicy>
	</samlp:AuthnRequest>
	`, id, time.Now().Format(time.RFC3339), assertionConsumerServiceURL, appIdUri)

	var buffer bytes.Buffer

	flateWriter, _ := flate.NewWriter(&buffer, -1)
	defer func(flateWriter *flate.Writer) {
		err := flateWriter.Close()
		if err != nil {
			log.Println(err)
		}
	}(flateWriter)

	_, err := flateWriter.Write([]byte(samlRequest))
	if err != nil {
		return "", err
	}

	err = flateWriter.Flush()
	if err != nil {
		return "", err
	}

	samlBase64 := base64.StdEncoding.EncodeToString(buffer.Bytes())

	return fmt.Sprintf("https://login.microsoftonline.com/%s/saml2?SAMLRequest=%s", tenantId, url.QueryEscape(samlBase64)), nil
}

func parseRolesFromSamlResponse(assertion string) ([]role, error) {
	b64, err := base64.StdEncoding.DecodeString(assertion)

	if err != nil {
		return nil, fmt.Errorf("failed to parse roles: %v", err)
	}

	var roles []role
	var sResponse samlResponse

	err = xml.Unmarshal(b64, &sResponse)

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal roles: %v", err)
	}

	for _, attr := range sResponse.Assertion.AttributeStatement.Attributes {
		if attr.Name == "https://aws.amazon.com/SAML/Attributes/Role" {
			for _, val := range attr.AttributeValues {
				parts := strings.Split(val.Value, ",")

				if strings.Contains(parts[0], ":role/") {
					roles = append(roles, role{
						roleArn:      strings.TrimSpace(parts[0]),
						principalArn: strings.TrimSpace(parts[1]),
					})
				} else {
					roles = append(roles, role{
						roleArn:      strings.TrimSpace(parts[1]),
						principalArn: strings.TrimSpace(parts[0]),
					})
				}

			}
		}
	}

	return roles, nil
}

func loginToJumpRole(config *configuration, saml string) (*role, *jumpRoleCredentials, error) {
	roles, err := parseRolesFromSamlResponse(saml)
	if err != nil {
		return nil, nil, err
	}

	rl, err := getJumpRole(roles, config, err)
	if err != nil {
		return nil, nil, err
	}

	durationSeconds := int32(config.DefaultDurationHours * 3600)

	jumpRoleStsInput := sts.AssumeRoleWithSAMLInput{
		PrincipalArn:    &rl.principalArn,
		RoleArn:         &rl.roleArn,
		SAMLAssertion:   &saml,
		DurationSeconds: &durationSeconds,
	}

	awsConfig, err := cfg.LoadDefaultConfig(context.Background())
	if err != nil {
		fmt.Println("Couldn't find the aws config for the specified profile. Creating a new one")
		awsConfig = *aws.NewConfig()
	}
	if awsConfig.Region == "" {
		awsConfig.Region = config.Region
	}

	jumpRoleClient := sts.NewFromConfig(awsConfig)
	jumpRoleResult, err := jumpRoleClient.AssumeRoleWithSAML(context.Background(), &jumpRoleStsInput)
	if err != nil {
		return nil, nil, err
	}

	return &rl, &jumpRoleCredentials{
		AwsAccessKeyId:     *jumpRoleResult.Credentials.AccessKeyId,
		AwsSecretAccessKey: *jumpRoleResult.Credentials.SecretAccessKey,
		AwsSessionToken:    *jumpRoleResult.Credentials.SessionToken,
		AwsExpiration:      *jumpRoleResult.Credentials.Expiration,
	}, nil
}
