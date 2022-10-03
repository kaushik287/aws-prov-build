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
//grpc= Google remote procedure call
/*
package clients

import (
	"context"
	"encoding/json"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/terrajet/pkg/terraform"

	"github.com/crossplane-contrib/provider-jet-aws/apis/v1alpha1"
)

const (
	// error messages
	errNoProviderConfig     = "no providerConfigRef provided"
	errGetProviderConfig    = "cannot get referenced ProviderConfig"
	errTrackUsage           = "cannot track ProviderConfig usage"
	errExtractCredentials   = "cannot extract credentials"
	errUnmarshalCredentials = "cannot unmarshal aws credentials as JSON"
	region                  = "us-east-1"
	keyAccessKeyID          = "access_key"
	keySecretAccessKey      = "secret_key"
)

// TerraformSetupBuilder builds Terraform a terraform.SetupFn function which
// returns Terraform provider setup configuration
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn {
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}
		//all good

		configRef := mg.GetProviderConfigReference()
		if configRef == nil {
			return ps, errors.New(errNoProviderConfig)
		}
		pc := &v1alpha1.ProviderConfig{}
		if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
			return ps, errors.Wrap(err, errGetProviderConfig)
		}

		t := resource.NewProviderConfigUsageTracker(client, &v1alpha1.ProviderConfigUsage{})
		if err := t.Track(ctx, mg); err != nil {
			return ps, errors.Wrap(err, errTrackUsage)
		}

		data, err := resource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, client, pc.Spec.Credentials.CommonCredentialSelectors)
		if err != nil {
			return ps, errors.Wrap(err, errExtractCredentials)
		}
		awsCreds := map[string]string{}
		if err := json.Unmarshal(data, &awsCreds); err != nil {
			return ps, errors.Wrap(err, errUnmarshalCredentials)
		}
/* donot uncomment
		// set environment variables for sensitive provider configuration
		// Deprecated: In shared gRPC mode we do not support injecting
		// credentials via the environment variables. You should specify
		// credentials via the Terraform main.tf.json instead.
		/*ps.Env = []string{
			fmt.Sprintf("%s=%s", "HASHICUPS_USERNAME", awsCreds["username"]),
			fmt.Sprintf("%s=%s", "HASHICUPS_PASSWORD", awsCreds["password"]),
		}*/
// set credentials in Terraform provider configuration
/*ps.Configuration = map[string]interface{}{
	"username": awsCreds["username"],
	"password": awsCreds["password"],
} do not uncomment*/
/*uncomment here
		ps.Configuration = map[string]interface{}{}
		if v, ok := awsCreds[keyAccessKeyID]; ok {
			ps.Configuration[keyAccessKeyID] = v
		}
		if v, ok := awsCreds[keySecretAccessKey]; ok {
			ps.Configuration[keySecretAccessKey] = v
		}
		if v, ok := awsCreds[region]; ok {
			ps.Configuration[region] = v
		}

		return ps, nil
	}
} */

package clients

import (
	"context"

	//"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/crossplane-contrib/provider-jet-aws/apis/v1alpha1"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/fieldpath"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	//xpabeta1 "github.com/crossplane/provider-aws/apis/v1beta1"
	//xpawsclient "github.com/crossplane/provider-aws/pkg/clients"
	"github.com/crossplane/terrajet/pkg/terraform"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// Terraform provider configuration keys for AWS credentials
	keySessionToken    = "token"
	keyAccessKeyID     = "access_key"
	keySecretAccessKey = "secret_key"
)

func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn { //nolint:gocyclo
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}
		if mg.GetProviderConfigReference() == nil {
			return ps, errors.New("no providerConfigRef provided")
		}
		pc := &v1alpha1.ProviderConfig{}
		if err := client.Get(ctx, types.NamespacedName{Name: mg.GetProviderConfigReference().Name}, pc); err != nil {
			return ps, errors.Wrap(err, "cannot get referenced Provider")
		}
		region, err := getRegion(mg)
		if err != nil {
			return ps, errors.Wrap(err, "cannot get region")
		}
		t := resource.NewProviderConfigUsageTracker(client, &v1alpha1.ProviderConfigUsage{})
		if err := t.Track(ctx, mg); err != nil {
			return ps, errors.Wrap(err, "cannot track ProviderConfig usage")
		}
		var cfg *aws.Config
		xpapc := &xpabeta1.ProviderConfig{
			Spec: xpabeta1.ProviderConfigSpec{
				Credentials:   xpabeta1.ProviderCredentials(pc.Spec.Credentials),
				AssumeRoleARN: pc.Spec.AssumeRoleARN,
			},
		}
		switch s := pc.Spec.Credentials.Source; s { //nolint:exhaustive
		case xpv1.CredentialsSourceInjectedIdentity:
			if pc.Spec.AssumeRoleARN != nil {
				if cfg, err = xpawsclient.UsePodServiceAccountAssumeRole(ctx, []byte{}, xpawsclient.DefaultSection, region, xpapc); err != nil {
					return ps, errors.Wrap(err, "failed to use pod service account assumeRoleARN")
				}
			} else {
				if cfg, err = xpawsclient.UsePodServiceAccount(ctx, []byte{}, xpawsclient.DefaultSection, region); err != nil {
					return ps, errors.Wrap(err, "failed to use pod service account")
				}
			}
		default:
			data, err := resource.CommonCredentialExtractor(ctx, s, client, pc.Spec.Credentials.CommonCredentialSelectors)
			if err != nil {
				return ps, errors.Wrap(err, "cannot get credentials")
			}
			if pc.Spec.AssumeRoleARN != nil {
				if cfg, err = xpawsclient.UseProviderSecretAssumeRole(ctx, data, xpawsclient.DefaultSection, region, xpapc); err != nil {
					return ps, errors.Wrap(err, "failed to use provider secret assumeRoleARN")
				}
			} else {
				if cfg, err = xpawsclient.UseProviderSecret(ctx, data, xpawsclient.DefaultSection, region); err != nil {
					return ps, errors.Wrap(err, "failed to use provider secret")
				}
			}
		}
		awsConf := xpawsclient.SetResolver(xpapc, cfg)
		creds, err := awsConf.Credentials.Retrieve(ctx)
		if err != nil {
			return ps, errors.Wrap(err, "failed to retrieve aws credentials from aws config")
		}
		// TODO(hasan): figure out what other values could be possible set here.
		//   e.g. what about setting an assume_role section: https://registry.terraform.io/providers/hashicorp/aws/latest/docs#argument-reference
		tfCfg := map[string]interface{}{}
		tfCfg["region"] = awsConf.Region
		if awsConf.Region == "" {
			// Some resources, like iam group, do not have a notion of region
			// hence we have no region in their schema. However, terraform still
			// attempts validating region in provider config and does not like
			// both empty string or not setting it at all. We need to skip
			// region validation in this case.
			tfCfg["skip_region_validation"] = true
		}
		// provider configuration for credentials
		tfCfg[keyAccessKeyID] = creds.AccessKeyID
		tfCfg[keySecretAccessKey] = creds.SecretAccessKey
		tfCfg[keySessionToken] = creds.SessionToken
		ps.Configuration = tfCfg
		return ps, err
	}
}

func getRegion(obj runtime.Object) (string, error) {
	fromMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return "", errors.Wrap(err, "cannot convert to unstructured")
	}
	r, err := fieldpath.Pave(fromMap).GetString("spec.forProvider.region")
	if fieldpath.IsNotFound(err) {
		// Region is not required for all resources, e.g. resource in "iam"
		// group.
		return "", nil
	}
	return r, err
}
