/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

/* additional tag key called tf in order to store the field name used in Terraform schema so that conversions
don't require strong-typed functions to be generated. Namely, the following mechanisms will be used for each
tf is a tag for translational function from terraform
tf in order to store the field name used in Terraform schema,
 so that conversions don't require strong-typed functions to be generated.
https://github.com/crossplane/crossplane/blob/master/design/design-doc-terrajet.md*/
import (
	// Note(turkenh): we are importing this to embed provider schema document
	_ "embed"

	tjconfig "github.com/crossplane/terrajet/pkg/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/crossplane-contrib/provider-jet-aws/config/awsdbinstance"
	"github.com/crossplane-contrib/provider-jet-aws/config/awss3"
	"github.com/crossplane-contrib/provider-jet-aws/config/iamuser"
)

const (
	resourcePrefix = "aws"
	modulePath     = "github.com/crossplane-contrib/provider-jet-aws"
)

//go:embed schema.json
var providerSchema string

// GetProvider returns provider configuration
//func function_name(Parameter-list)(Return_type)

func GetProvider() *tjconfig.Provider {
	//tjconfig.Provider is a variable with return type here of pointer
	//pointer can be returned as variables in go
	//DefaultResourceFn returns a default resource configuration to be used while building resource configurations.
	//DefaultResource keeps an initial default configuration for all resources of a provider.
	defaultResourceFn := func(name string, terraformResource *schema.Resource, opts ...tjconfig.ResourceOption) *tjconfig.Resource {
		r := tjconfig.DefaultResource(name, terraformResource)
		// Add any provider-specific defaulting here. For example:
		//   r.ExternalName = tjconfig.IdentifierFromProvider
		return r
	}
	//NewProviderWithSchema builds and returns a new Provider from provider tfjson schema,
	// that is generated using Terraform CLI with: `terraform providers schema --json`
	//refer make file
	pc := tjconfig.NewProviderWithSchema([]byte(providerSchema), resourcePrefix, modulePath,
		tjconfig.WithDefaultResourceFn(defaultResourceFn),
		//WithDefaultResourceFn configures DefaultResourceFn for this Provider
		//func WithIncludeList(l []string) ProviderOption
		//provider option is used to configure a provider
		//WithIncludeList configures IncludeList for this Provider.
		tjconfig.WithIncludeList([]string{
			"aws_iam_user$",
			"aws_db_instance$",
			"aws_s3_bucket$",
		}))

	//for _, o := range opts {
	//o(p)
	//}
	for _, configure := range []func(provider *tjconfig.Provider){
		// add custom config functions
		// above in func(provider tjconfig.Provider) call by reference is done
		// so what ever the value present in tjconfig.Provider address is been called and here for every resource Configure
		//function is executed
		//check import section for tjconfig addresses range
		iamuser.Configure,
		awsdbinstance.Configure,
		awss3.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()
	return pc
}

// ConfigureResources configures resources with provided ResourceConfigurator's -->config.go files for each resource
//func (p *Provider) ConfigureResources() {}
