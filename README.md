Golang Marshaler from Environment Variables
===========================================

[![Build Status](https://travis-ci.org/evilwire/go-env.svg?branch=master)](https://travis-ci.org/evilwire/go-env)
[![GoDoc](https://godoc.org/github.com/evilwire/go-env?status.svg)](https://godoc.org/github.com/evilwire/go-env)
[![Go Report Card](https://goreportcard.com/badge/github.com/evilwire/go-env)](https://goreportcard.com/report/github.com/evilwire/go-env)
[![codecov](https://codecov.io/gh/evilwire/go-env/branch/master/graph/badge.svg)](https://codecov.io/gh/evilwire/go-env)

Golang (1.5+) package for marshalling objects from environment variable values.
There are many like packages. The one that inspires this API the most is
the `json` package. The idea is that configuration objects are stored
as environment variables, especially when running as a containerised
application.

### Example

The following is a full and well contrived application that showcases
how to use the API to retrieve data from environments. The application

- retrieves a list of things to say and wait time (via config)
- for everything to be said, it prints a simple statement

Disclaimer: this is a silly toy application. Don't use this as an example
of good application design. It isn't.

```go
package main


import (
    "fmt"
    "github.com/evilwire/go-env"
    "time"
)


type SomeConfig struct {
        User string `env:"USER_NAME"`
        
        Application struct {
                WaitDuration time.Duration `env:"WAIT_DURATION"`
                
                // this is best obtained from flags, but you can equally
                // retrieve this from environment variables
                ThingsToSay []string `env:"THINGS_TO_SAY"`
        } `env:"APP_"`
}


// Sets up my application by reading the configs
func setup(env goenv.EnvReader) (*SomeConfig, error) {
        // create a Marshaler using our lovely default, which knows
        // how to marshal a set of things
        marshaller := goenv.DefaultEnvMarshaler{
                env,
        }
        
        // instantiate an empty config 
        config := SomeConfig{} 
        err := marshaller.Unmarshal(&config)
        if err != nil {
            return nil, err
        }
        
        return &config, nil
}


func main() {
        config, err := setup(goenv.NewOsEnvReader())
        if err != nil {
                panic(err)
        }
        
        for _, line := range config.Application.ThingsToSay {
                fmt.Printf("%s says: ", config.User)
                fmt.Println(line)
                time.Sleep(config.Application.WaitDuration)
        }
        
        fmt.Println("We're done!")
}

```

Compile your application, say `silly-app`, and run your application

```sh
USER_NAME='Michael Bluth' \
APP_WAIT_DURATION=2s \
APP_THINGS_TO_SAY='hiya,how are you,bye' /path/to/silly-app
```

### Customising the `UnmarshalEnv` method

One of the cases that I encounter is using [AWS KMS](https://aws.amazon.com/kms/) to manage
secrets. For example, you may want to store KMS-encrypted credentials in a distributed 
version control source such GitHub, and passed into your application directly via 
environment variables.

We want to use the following function to decrypt KMS-encrypted secrets
```go
package whatever

import (
        "github.com/aws/aws-sdk-go/service/kms"
        "github.com/aws/aws-sdk-go/service/kms/kmsiface"
        "encoding/base64"
)


// Uses KMS to decrypt an encrypted, base64 encoded secret (string) 
// by base64 decoding and KMS decrypting the bugger
func KMSDecrypt(secret string, kmsClient kmsiface.KMSAPI) (string, error) {
        b64Secret, err := base64.StdEncoding.DecodeString(secret)
        if err != nil {
                // handle
                return "", err
        }
        response, err := kmsClient.Decrypt(&kms.DecryptInput {
                CiphertextBlob: b64Secret,
                
                // additional context???
        })
        if err != nil {
            return "", err
        }
        
        return string(response.Plaintext), nil
}
```

Let's make a custom marshal method that ingests KMS-encrypted

```go
package whatever

import (
    "github.com/evilwire/go-env"
    "github.com/aws/aws-sdk-go/service/kms"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/aws"
)

type KMSEncryptedConfig struct {
        Username string `env:"USER"`
        Password string
        
        // and other fields that one might be able
        // to get from things
}


func (config *KMSEncryptedConfig) MarshalEnv(env goenv.EnvReader) error {
        tempConfig := struct {
                KMSEncryptedConfig
                KMSPassword string `env:"KMS_PASSWORD"`
        }{}
        
        marshaller := goenv.DefaultEnvMarshaler{ env }
        err := marshaller.Unmarshal(&tempConfig)
        if err != nil {
                return err
        }
        
        password, err := KMSDecrypt(tempConfig.KMSPassword, &kms.New(session.New(aws.Config{
            // configuration of some sort
        })))
        if err != nil {
                return err
        }
        
        config.Username = tempConfig.Username
        config.Password = password
        // more copying...
        
        return nil
}

```

Now you can write an application and accepts KMS passwords and not have to worry:

```go
package main


import (
        "whatever"
        "github.com/evilwire/go-env"
)


type Config struct {
        DbCredentials *whatever.KMSEncryptedConfig `env:"DB_"`
        
        // other types of configs
}


func doStuff(config *Config) error {
        // do stuff
        
        return nil
}


func setup(env goenv.EnvReader) (*Config, error) {
        config := Config {}
        marshaller := goenv.DefaultEnvMarshaler{ env }
        
        err := marshaller.Unmarshal(&config)
        if err != nil {
                return nil, err
        }
        
        // does setuppy things...
        
        return config, nil
}


func main() {
        config, err := setup(goenv.NewOsEnvReader())
        if err != nil {
                panic(err)
        }
        
        // does stuff with that config
        panic(doStuff(config))
}
```

Now compile your application as, say `app`, and run the application as:

```bash
DB_KMS_PASSWORD=ABCD1234abcd1234 \
DB_USER=mbluth \
#... \
app
```
