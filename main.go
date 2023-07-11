package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	client "github.com/cloudbase/garm/client"
	clientInstances "github.com/cloudbase/garm/client/instances"
	clientOrganizations "github.com/cloudbase/garm/client/organizations"
	clientRepositories "github.com/cloudbase/garm/client/repositories"
	"github.com/cloudbase/garm/cmd/garm-cli/config"
	"github.com/cloudbase/garm/params"
	"github.com/cloudbase/garm/runner/providers/common"
	"github.com/go-openapi/runtime"
	openapiRuntimeClient "github.com/go-openapi/runtime/client"
)

const (
	orgName  = "test-garm-org"
	repoName = "test-garm-repo"
)

var (
	cli       *client.GarmAPI
	authToken runtime.ClientAuthInfoWriter

	credentialsName = os.Getenv("CREDENTIALS_NAME")

	repoID            string
	repoPoolID        string
	repoInstanceName  string
	repoWebhookSecret = os.Getenv("REPO_WEBHOOK_SECRET")

	orgID            string
	orgWebhookSecret = os.Getenv("ORG_WEBHOOK_SECRET")

	instanceName string
)

func handleError(err error) {
	if err != nil {
		log.Fatalf("error encountered: %v", err)
	}
}

func printResponse(resp interface{}) {
	b, err := json.MarshalIndent(resp, "", "  ")
	handleError(err)
	log.Println(string(b))
}

// Repository
func CreateRepo() {
	listReposResp, err := cli.Repositories.ListRepos(
		clientRepositories.NewListReposParams(),
		authToken)
	handleError(err)
	if len(listReposResp.Payload) > 0 {
		log.Println(">>> Repo already exists, skipping create")
		repoID = listReposResp.Payload[0].ID
		return
	}
	log.Println(">>> Create repo")
	createRepoResp, err := cli.Repositories.CreateRepo(
		clientRepositories.NewCreateRepoParams().
			WithBody(
				params.CreateRepoParams{
					Owner:           orgName,
					Name:            repoName,
					CredentialsName: credentialsName,
					WebhookSecret:   repoWebhookSecret,
				}),
		authToken)
	handleError(err)
	printResponse(createRepoResp.Payload)
	repoID = createRepoResp.Payload.ID
}

func ListRepos() {
	log.Println(">>> List repos")
	listReposResp, err := cli.Repositories.ListRepos(
		clientRepositories.NewListReposParams(),
		authToken)
	handleError(err)
	printResponse(listReposResp.Payload)
	repoID = listReposResp.Payload[0].ID
}

func UpdateRepo() {
	log.Println(">>> Update repo")
	updateRepoResp, err := cli.Repositories.UpdateRepo(
		clientRepositories.NewUpdateRepoParams().
			WithRepoID(repoID).
			WithBody(
				params.UpdateEntityParams{
					CredentialsName: fmt.Sprintf("%s-clone", credentialsName),
				}),
		authToken)
	handleError(err)
	printResponse(updateRepoResp.Payload)
}

func GetRepo() {
	log.Println(">>> Get repo")
	getRepoResp, err := cli.Repositories.GetRepo(
		clientRepositories.NewGetRepoParams().
			WithRepoID(repoID),
		authToken)
	handleError(err)
	printResponse(getRepoResp.Payload)
}

func CreateRepoPool() {
	listRepoPoolsResp, err := cli.Repositories.ListRepoPools(
		clientRepositories.NewListRepoPoolsParams().
			WithRepoID(repoID),
		authToken)
	handleError(err)
	if len(listRepoPoolsResp.Payload) > 0 {
		log.Println(">>> Repo pool already exists, skipping create")
		repoPoolID = listRepoPoolsResp.Payload[0].ID
		return
	}
	log.Println(">>> Create repo pool")
	createRepoPoolResp, err := cli.Repositories.CreateRepoPool(
		clientRepositories.NewCreateRepoPoolParams().
			WithRepoID(repoID).
			WithBody(params.CreatePoolParams{
				MaxRunners:     2,
				MinIdleRunners: 0,
				Flavor:         "garm",
				Image:          "ubuntu:22.04",
				OSType:         params.Linux,
				OSArch:         params.Amd64,
				ProviderName:   "lxd_local",
				Tags:           []string{"ubuntu", "simple-runner", "repo-runner"},
				Enabled:        true,
			}),
		authToken)
	handleError(err)
	printResponse(createRepoPoolResp.Payload)
	repoPoolID = createRepoPoolResp.Payload.ID
}

func ListRepoPools() {
	log.Println(">>> List repo pools")
	listRepoPoolsResp, err := cli.Repositories.ListRepoPools(
		clientRepositories.NewListRepoPoolsParams().
			WithRepoID(repoID),
		authToken)
	handleError(err)
	printResponse(listRepoPoolsResp.Payload)
}

func GetRepoPool() {
	log.Println(">>> Get repo pool")
	getRepoPoolResp, err := cli.Repositories.GetRepoPool(
		clientRepositories.NewGetRepoPoolParams().
			WithRepoID(repoID).
			WithPoolID(repoPoolID),
		authToken)
	handleError(err)
	printResponse(getRepoPoolResp.Payload)
}

func UpdateRepoPool() {
	log.Println(">>> Update repo pool")
	var maxRunners uint = 5
	var idleRunners uint = 1
	updateRepoPoolResp, err := cli.Repositories.UpdateRepoPool(
		clientRepositories.NewUpdateRepoPoolParams().
			WithRepoID(repoID).
			WithPoolID(repoPoolID).
			WithBody(params.UpdatePoolParams{
				MinIdleRunners: &idleRunners,
				MaxRunners:     &maxRunners,
			}),
		authToken)
	handleError(err)
	printResponse(updateRepoPoolResp.Payload)
}

func DisableRepoPool() {
	enabled := false
	_, err := cli.Repositories.UpdateRepoPool(
		clientRepositories.NewUpdateRepoPoolParams().
			WithRepoID(repoID).
			WithPoolID(repoPoolID).
			WithBody(params.UpdatePoolParams{
				Enabled: &enabled,
			}),
		authToken)
	handleError(err)
	log.Printf("repo pool %s disabled", repoPoolID)
}

func WaitRepoPoolNoInstances() {
	for {
		log.Println(">>> Wait until repo pool has no instances")
		getRepoPoolResp, err := cli.Repositories.GetRepoPool(
			clientRepositories.NewGetRepoPoolParams().
				WithRepoID(repoID).
				WithPoolID(repoPoolID),
			authToken)
		handleError(err)
		if len(getRepoPoolResp.Payload.Instances) == 0 {
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func WaitRepoInstance() {
	log.Println(">>> Wait until repo instance is in running state")
	for {
		listRepoInstancesResp, err := cli.Repositories.ListRepoInstances(
			clientRepositories.NewListRepoInstancesParams().
				WithRepoID(repoID),
			authToken)
		handleError(err)
		if len(listRepoInstancesResp.Payload) > 0 {
			instance := listRepoInstancesResp.Payload[0]
			log.Printf("instance %s status: %s", instance.Name, instance.Status)
			if instance.Status == common.InstanceRunning && instance.RunnerStatus == common.RunnerIdle {
				repoInstanceName = instance.Name
				break
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func ListRepoInstances() {
	log.Println(">>> List repo instances")
	listRepoInstancesResp, err := cli.Repositories.ListRepoInstances(
		clientRepositories.NewListRepoInstancesParams().
			WithRepoID(repoID),
		authToken)
	handleError(err)
	printResponse(listRepoInstancesResp.Payload)
}

// Organizations
func CreateOrg() {
	listOrgsResp, err := cli.Organizations.ListOrgs(
		clientOrganizations.NewListOrgsParams(),
		authToken)
	handleError(err)
	if len(listOrgsResp.Payload) > 0 {
		log.Println(">>> Org already exists, skipping create")
		orgID = listOrgsResp.Payload[0].ID
		return
	}
	log.Println(">>> Create org")
	createOrgResp, err := cli.Organizations.CreateOrg(
		clientOrganizations.NewCreateOrgParams().
			WithBody(
				params.CreateOrgParams{
					Name:            orgName,
					CredentialsName: credentialsName,
					WebhookSecret:   orgWebhookSecret,
				}),
		authToken)
	handleError(err)
	printResponse(createOrgResp.Payload)
	orgID = createOrgResp.Payload.ID
}

func ListOrgs() {
	log.Println(">>> List orgs")
	listOrgsResp, err := cli.Organizations.ListOrgs(
		clientOrganizations.NewListOrgsParams(),
		authToken)
	handleError(err)
	printResponse(listOrgsResp.Payload)
	orgID = listOrgsResp.Payload[0].ID
}

func UpdateOrg() {
	log.Println(">>> Update org")
	updateOrgResp, err := cli.Organizations.UpdateOrg(
		clientOrganizations.NewUpdateOrgParams().
			WithOrgID(orgID).
			WithBody(
				params.UpdateEntityParams{
					CredentialsName: fmt.Sprintf("%s-clone", credentialsName),
				}),
		authToken)
	handleError(err)
	printResponse(updateOrgResp.Payload)
}

func GetOrg() {
	log.Println(">>> Get org")
	getOrgResp, err := cli.Organizations.GetOrg(
		clientOrganizations.NewGetOrgParams().
			WithOrgID(orgID),
		authToken)
	handleError(err)
	printResponse(getOrgResp.Payload)
}

// Instances
func ListInstances() {
	log.Println(">>> List instances")
	listInstancesResp, err := cli.Instances.ListInstances(
		clientInstances.NewListInstancesParams(),
		authToken)
	handleError(err)
	printResponse(listInstancesResp.Payload)
	instanceName = listInstancesResp.Payload[0].Name
}

func GetInstance() {
	log.Println(">>> Get instance")
	getInstanceResp, err := cli.Instances.GetInstance(
		clientInstances.NewGetInstanceParams().
			WithInstanceName(instanceName),
		authToken)
	handleError(err)
	printResponse(getInstanceResp.Payload)
}

func DeleteInstance(name string) {
	err := cli.Instances.DeleteInstance(
		clientInstances.NewDeleteInstanceParams().
			WithInstanceName(name),
		authToken)
	for {
		log.Printf(">>> Wait until instance %s is deleted", name)
		listInstancesResp, err := cli.Instances.ListInstances(
			clientInstances.NewListInstancesParams(),
			authToken)
		handleError(err)
		for _, instance := range listInstancesResp.Payload {
			if instance.Name == name {
				time.Sleep(5 * time.Second)
				continue
			}
		}
		break
	}
	handleError(err)
	log.Printf("instance %s deleted", name)
}

func DeleteRepoPool() {
	log.Println(">>> Delete repo pool")
	err := cli.Repositories.DeleteRepoPool(
		clientRepositories.NewDeleteRepoPoolParams().
			WithRepoID(repoID).
			WithPoolID(repoPoolID),
		authToken)
	handleError(err)
	log.Printf("repo pool %s deleted", repoPoolID)
}

func DeleteRepo() {
	log.Println(">>> Delete repo")
	err := cli.Repositories.DeleteRepo(
		clientRepositories.NewDeleteRepoParams().
			WithRepoID(repoID),
		authToken)
	handleError(err)
	log.Printf("repo %s deleted", repoID)
}

func DeleteOrg() {
	log.Println(">>> Delete org")
	err := cli.Organizations.DeleteOrg(
		clientOrganizations.NewDeleteOrgParams().
			WithOrgID(orgID),
		authToken)
	handleError(err)
	log.Printf("org %s deleted", orgID)
}

func main() {
	//
	// Load GARM client config
	//
	cfg, err := config.LoadConfig()
	handleError(err)
	garmUrl, err := url.Parse(cfg.Managers[0].BaseURL)
	handleError(err)
	authToken = openapiRuntimeClient.BearerToken(cfg.Managers[0].Token)
	apiPath, err := url.JoinPath(garmUrl.Path, client.DefaultBasePath)
	handleError(err)

	//
	// Create GARM client
	//
	transportCfg := client.DefaultTransportConfig().
		WithHost(garmUrl.Host).
		WithBasePath(apiPath).
		WithSchemes([]string{garmUrl.Scheme})
	cli = client.NewHTTPClientWithConfig(nil, transportCfg)

	//////////////////
	// repositories //
	//////////////////
	CreateRepo()
	ListRepos()
	UpdateRepo()
	GetRepo()

	CreateRepoPool()
	ListRepoPools()
	GetRepoPool()
	UpdateRepoPool()
	WaitRepoInstance()

	ListRepoInstances()

	//////////////////
	// organizations //
	//////////////////
	CreateOrg()
	ListOrgs()
	UpdateOrg()
	GetOrg()

	///////////////
	// instances //
	///////////////
	ListInstances()
	GetInstance()

	/////////////
	// Cleanup //
	/////////////
	DisableRepoPool()
	DeleteInstance(repoInstanceName)
	WaitRepoPoolNoInstances()
	DeleteRepoPool()
	DeleteRepo()
	DeleteOrg()
}
