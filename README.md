# Awsure

If your organization utilizes Azure Active Directory for single sign-on (SSO) authentication to access the AWS console, 
logging in through the command line or using the AWS CLI isn't straightforward. This tool addresses that issue by enabling 
you to employ the standard Azure AD login process (including multi-factor authentication) from the command line. 
It establishes a federated AWS session and stores the temporary credentials in the correct location for the AWS CLI and SDKs to utilize. 

## Installation

Download the executable from the release page for your operating system, or if you have `go` installed, run:
```shell
go install github.com/vahid-haghighat/awsure@latest
```

## Acknowledgment
This project is based on [go-aws-azure-login](https://github.com/luneo7/go-aws-azure-login).
