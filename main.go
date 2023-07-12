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
	clientPools "github.com/cloudbase/garm/client/pools"
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
	orgPoolID        string
	orgInstanceName  string
	orgWebhookSecret = os.Getenv("ORG_WEBHOOK_SECRET")

	instanceName string
	poolID       string
)

// //////////////// //
// helper functions //
// ///////////////////
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

// ///////////////
// Repositories //
// ///////////////
func createRepo(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter, repoParams params.CreateRepoParams) (*params.Repository, error) {
	createRepoResponse, err := apiCli.Repositories.CreateRepo(
		clientRepositories.NewCreateRepoParams().WithBody(repoParams),
		apiAuthToken)
	if err != nil {
		return nil, err
	}
	return &createRepoResponse.Payload, nil
}

func listRepos(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter) (params.Repositories, error) {
	listReposResponse, err := apiCli.Repositories.ListRepos(
		clientRepositories.NewListReposParams(),
		apiAuthToken)
	if err != nil {
		return nil, err
	}
	return listReposResponse.Payload, nil
}

func updateRepo(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter, repoID string, repoParams params.UpdateEntityParams) (*params.Repository, error) {
	updateRepoResponse, err := apiCli.Repositories.UpdateRepo(
		clientRepositories.NewUpdateRepoParams().WithRepoID(repoID).WithBody(repoParams),
		apiAuthToken)
	if err != nil {
		return nil, err
	}
	return &updateRepoResponse.Payload, nil
}

func getRepo(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter, repoID string) (*params.Repository, error) {
	getRepoResponse, err := apiCli.Repositories.GetRepo(
		clientRepositories.NewGetRepoParams().WithRepoID(repoID),
		apiAuthToken)
	if err != nil {
		return nil, err
	}
	return &getRepoResponse.Payload, nil
}

func createRepoPool(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter, repoID string, poolParams params.CreatePoolParams) (*params.Pool, error) {
	createRepoPoolResponse, err := apiCli.Repositories.CreateRepoPool(
		clientRepositories.NewCreateRepoPoolParams().WithRepoID(repoID).WithBody(poolParams),
		apiAuthToken)
	if err != nil {
		return nil, err
	}
	return &createRepoPoolResponse.Payload, nil
}

func listRepoPools(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter, repoID string) (params.Pools, error) {
	listRepoPoolsResponse, err := apiCli.Repositories.ListRepoPools(
		clientRepositories.NewListRepoPoolsParams().WithRepoID(repoID),
		apiAuthToken)
	if err != nil {
		return nil, err
	}
	return listRepoPoolsResponse.Payload, nil
}

func getRepoPool(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter, repoID, poolID string) (*params.Pool, error) {
	getRepoPoolResponse, err := apiCli.Repositories.GetRepoPool(
		clientRepositories.NewGetRepoPoolParams().WithRepoID(repoID).WithPoolID(poolID),
		apiAuthToken)
	if err != nil {
		return nil, err
	}
	return &getRepoPoolResponse.Payload, nil
}

func updateRepoPool(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter, repoID, poolID string, poolParams params.UpdatePoolParams) (*params.Pool, error) {
	updateRepoPoolResponse, err := apiCli.Repositories.UpdateRepoPool(
		clientRepositories.NewUpdateRepoPoolParams().WithRepoID(repoID).WithPoolID(poolID).WithBody(poolParams),
		apiAuthToken)
	if err != nil {
		return nil, err
	}
	return &updateRepoPoolResponse.Payload, nil
}

func listRepoInstances(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter, repoID string) (params.Instances, error) {
	listRepoInstancesResponse, err := apiCli.Repositories.ListRepoInstances(
		clientRepositories.NewListRepoInstancesParams().WithRepoID(repoID),
		apiAuthToken)
	if err != nil {
		return nil, err
	}
	return listRepoInstancesResponse.Payload, nil
}

func deleteRepo(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter, repoID string) error {
	return apiCli.Repositories.DeleteRepo(
		clientRepositories.NewDeleteRepoParams().WithRepoID(repoID),
		apiAuthToken)
}

// ////////////////
// Organizations //
// ////////////////
func createOrg(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter, orgParams params.CreateOrgParams) (*params.Organization, error) {
	createOrgResponse, err := apiCli.Organizations.CreateOrg(
		clientOrganizations.NewCreateOrgParams().WithBody(orgParams),
		apiAuthToken)
	if err != nil {
		return nil, err
	}
	return &createOrgResponse.Payload, nil
}

func listOrgs(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter) (params.Organizations, error) {
	listOrgsResponse, err := apiCli.Organizations.ListOrgs(
		clientOrganizations.NewListOrgsParams(),
		apiAuthToken)
	if err != nil {
		return nil, err
	}
	return listOrgsResponse.Payload, nil
}

// ////////
// Pools //
// ////////
func listPools(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter) (params.Pools, error) {
	listPoolsResponse, err := apiCli.Pools.ListPools(
		clientPools.NewListPoolsParams(),
		apiAuthToken)
	if err != nil {
		return nil, err
	}
	return listPoolsResponse.Payload, nil
}

func getPool(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter, poolID string) (*params.Pool, error) {
	getPoolResponse, err := apiCli.Pools.GetPool(
		clientPools.NewGetPoolParams().WithPoolID(poolID),
		apiAuthToken)
	if err != nil {
		return nil, err
	}
	return &getPoolResponse.Payload, nil
}

func deletePool(apiCli *client.GarmAPI, apiAuthToken runtime.ClientAuthInfoWriter, poolID string) error {
	return apiCli.Pools.DeletePool(
		clientPools.NewDeletePoolParams().WithPoolID(poolID),
		apiAuthToken)
}

// /////////////////
// Main functions //
// /////////////////
//
// ///////////////
// Repositories //
// ///////////////
func CreateRepo() {
	repos, err := listRepos(cli, authToken)
	handleError(err)
	if len(repos) > 0 {
		log.Println(">>> Repo already exists, skipping create")
		repoID = repos[0].ID
		return
	}
	log.Println(">>> Create repo")
	createParams := params.CreateRepoParams{
		Owner:           orgName,
		Name:            repoName,
		CredentialsName: credentialsName,
		WebhookSecret:   repoWebhookSecret,
	}
	repo, err := createRepo(cli, authToken, createParams)
	handleError(err)
	printResponse(repo)
	repoID = repo.ID
}

func ListRepos() {
	log.Println(">>> List repos")
	repos, err := listRepos(cli, authToken)
	handleError(err)
	printResponse(repos)
}

func UpdateRepo() {
	log.Println(">>> Update repo")
	updateParams := params.UpdateEntityParams{
		CredentialsName: fmt.Sprintf("%s-clone", credentialsName),
	}
	repo, err := updateRepo(cli, authToken, repoID, updateParams)
	handleError(err)
	printResponse(repo)
}

func GetRepo() {
	log.Println(">>> Get repo")
	repo, err := getRepo(cli, authToken, repoID)
	handleError(err)
	printResponse(repo)
}

func CreateRepoPool() {
	pools, err := listRepoPools(cli, authToken, repoID)
	handleError(err)
	if len(pools) > 0 {
		log.Println(">>> Repo pool already exists, skipping create")
		repoPoolID = pools[0].ID
		return
	}
	log.Println(">>> Create repo pool")
	poolParams := params.CreatePoolParams{
		MaxRunners:     2,
		MinIdleRunners: 0,
		Flavor:         "garm",
		Image:          "ubuntu:22.04",
		OSType:         params.Linux,
		OSArch:         params.Amd64,
		ProviderName:   "lxd_local",
		Tags:           []string{"ubuntu", "simple-runner"},
		Enabled:        true,
	}
	repo, err := createRepoPool(cli, authToken, repoID, poolParams)
	handleError(err)
	printResponse(repo)
	repoPoolID = repo.ID
}

func ListRepoPools() {
	log.Println(">>> List repo pools")
	pools, err := listRepoPools(cli, authToken, repoID)
	handleError(err)
	printResponse(pools)
}

func GetRepoPool() {
	log.Println(">>> Get repo pool")
	pool, err := getRepoPool(cli, authToken, repoID, repoPoolID)
	handleError(err)
	printResponse(pool)
}

func UpdateRepoPool() {
	log.Println(">>> Update repo pool")
	var maxRunners uint = 5
	var idleRunners uint = 1
	poolParams := params.UpdatePoolParams{
		MinIdleRunners: &idleRunners,
		MaxRunners:     &maxRunners,
	}
	pool, err := updateRepoPool(cli, authToken, repoID, repoPoolID, poolParams)
	handleError(err)
	printResponse(pool)
}

func DisableRepoPool() {
	enabled := false
	_, err := updateRepoPool(cli, authToken, repoID, repoPoolID, params.UpdatePoolParams{Enabled: &enabled})
	handleError(err)
	log.Printf("repo pool %s disabled", repoPoolID)
}

func WaitRepoPoolNoInstances() {
	for {
		log.Println(">>> Wait until repo pool has no instances")
		pool, err := getRepoPool(cli, authToken, repoID, repoPoolID)
		handleError(err)
		if len(pool.Instances) == 0 {
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func WaitRepoInstance() {
	log.Println(">>> Wait until repo instance is in running state")
	for {
		instances, err := listRepoInstances(cli, authToken, repoID)
		handleError(err)
		if len(instances) > 0 {
			instance := instances[0]
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
	instances, err := listRepoInstances(cli, authToken, repoID)
	handleError(err)
	printResponse(instances)
}

func DeleteRepo() {
	log.Println(">>> Delete repo")
	err := deleteRepo(cli, authToken, repoID)
	handleError(err)
	log.Printf("repo %s deleted", repoID)
}

// TODO only
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

// ////////////////
// Organizations //
// ////////////////
func CreateOrg() {
	orgs, err := listOrgs(cli, authToken)
	handleError(err)
	if len(orgs) > 0 {
		log.Println(">>> Org already exists, skipping create")
		orgID = orgs[0].ID
		return
	}
	log.Println(">>> Create org")
	orgParams := params.CreateOrgParams{
		Name:            orgName,
		CredentialsName: credentialsName,
		WebhookSecret:   orgWebhookSecret,
	}
	org, err := createOrg(cli, authToken, orgParams)
	handleError(err)
	printResponse(org)
	orgID = org.ID
}

func ListOrgs() {
	log.Println(">>> List orgs")
	orgs, err := listOrgs(cli, authToken)
	handleError(err)
	printResponse(orgs)
}

// TODO below
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

func CreateOrgPool() {
	listOrgPoolsResp, err := cli.Organizations.ListOrgPools(
		clientOrganizations.NewListOrgPoolsParams().
			WithOrgID(orgID),
		authToken)
	handleError(err)
	if len(listOrgPoolsResp.Payload) > 0 {
		log.Println(">>> Org pool already exists, skipping create")
		orgPoolID = listOrgPoolsResp.Payload[0].ID
		return
	}
	log.Println(">>> Create org pool")
	createOrgPoolResp, err := cli.Organizations.CreateOrgPool(
		clientOrganizations.NewCreateOrgPoolParams().
			WithOrgID(orgID).
			WithBody(params.CreatePoolParams{
				MaxRunners:     2,
				MinIdleRunners: 0,
				Flavor:         "garm",
				Image:          "ubuntu:22.04",
				OSType:         params.Linux,
				OSArch:         params.Amd64,
				ProviderName:   "lxd_local",
				Tags:           []string{"ubuntu", "simple-runner", "org-runner"},
				Enabled:        true,
			}),
		authToken)
	handleError(err)
	printResponse(createOrgPoolResp.Payload)
	orgPoolID = createOrgPoolResp.Payload.ID
}

func ListOrgPools() {
	log.Println(">>> List org pools")
	listOrgPoolsResp, err := cli.Organizations.ListOrgPools(
		clientOrganizations.NewListOrgPoolsParams().
			WithOrgID(orgID),
		authToken)
	handleError(err)
	printResponse(listOrgPoolsResp.Payload)
}

func GetOrgPool() {
	log.Println(">>> Get org pool")
	getOrgPoolResp, err := cli.Organizations.GetOrgPool(
		clientOrganizations.NewGetOrgPoolParams().
			WithOrgID(orgID).
			WithPoolID(orgPoolID),
		authToken)
	handleError(err)
	printResponse(getOrgPoolResp.Payload)
}

func UpdateOrgPool() {
	log.Println(">>> Update org pool")
	var maxRunners uint = 5
	var idleRunners uint = 1
	updateOrgPoolResp, err := cli.Organizations.UpdateOrgPool(
		clientOrganizations.NewUpdateOrgPoolParams().
			WithOrgID(orgID).
			WithPoolID(orgPoolID).
			WithBody(params.UpdatePoolParams{
				MinIdleRunners: &idleRunners,
				MaxRunners:     &maxRunners,
			}),
		authToken)
	handleError(err)
	printResponse(updateOrgPoolResp.Payload)
}

func DisableOrgPool() {
	enabled := false
	_, err := cli.Organizations.UpdateOrgPool(
		clientOrganizations.NewUpdateOrgPoolParams().
			WithOrgID(orgID).
			WithPoolID(orgPoolID).
			WithBody(params.UpdatePoolParams{
				Enabled: &enabled,
			}),
		authToken)
	handleError(err)
	log.Printf("org pool %s disabled", orgPoolID)
}

func WaitOrgPoolNoInstances() {
	for {
		log.Println(">>> Wait until org pool has no instances")
		getOrgPoolResp, err := cli.Organizations.GetOrgPool(
			clientOrganizations.NewGetOrgPoolParams().
				WithOrgID(orgID).
				WithPoolID(orgPoolID),
			authToken)
		handleError(err)
		if len(getOrgPoolResp.Payload.Instances) == 0 {
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func WaitOrgInstance() {
	log.Println(">>> Wait until org instance is in running state")
	for {
		listOrgInstancesResp, err := cli.Organizations.ListOrgInstances(
			clientOrganizations.NewListOrgInstancesParams().
				WithOrgID(orgID),
			authToken)
		handleError(err)
		if len(listOrgInstancesResp.Payload) > 0 {
			instance := listOrgInstancesResp.Payload[0]
			log.Printf("instance %s status: %s", instance.Name, instance.Status)
			if instance.Status == common.InstanceRunning && instance.RunnerStatus == common.RunnerIdle {
				orgInstanceName = instance.Name
				break
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func ListOrgInstances() {
	log.Println(">>> List org instances")
	listOrgInstancesResp, err := cli.Organizations.ListOrgInstances(
		clientOrganizations.NewListOrgInstancesParams().
			WithOrgID(orgID),
		authToken)
	handleError(err)
	printResponse(listOrgInstancesResp.Payload)
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

func DeleteOrgPool() {
	log.Println(">>> Delete org pool")
	err := cli.Organizations.DeleteOrgPool(
		clientOrganizations.NewDeleteOrgPoolParams().
			WithOrgID(orgID).
			WithPoolID(orgPoolID),
		authToken)
	handleError(err)
	log.Printf("org pool %s deleted", orgPoolID)
}

// ////////////
// Instances //
// ////////////
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

// ////////
// Pools //
// ////////
func CreatePool() {
	pools, err := listPools(cli, authToken)
	handleError(err)
	for _, pool := range pools {
		if pool.Image == "ubuntu:20.04" {
			// this is the extra pool to be deleted, later, via [DELETE] pools dedicated API.
			poolID = pool.ID
			return
		}
	}
	log.Println(">>> Create pool")
	poolParams := params.CreatePoolParams{
		MaxRunners:     2,
		MinIdleRunners: 0,
		Flavor:         "garm",
		Image:          "ubuntu:20.04",
		OSType:         params.Linux,
		OSArch:         params.Amd64,
		ProviderName:   "lxd_local",
		Tags:           []string{"ubuntu", "simple-runner"},
		Enabled:        true,
	}
	pool, err := createRepoPool(cli, authToken, repoID, poolParams)
	handleError(err)
	printResponse(pool)
	poolID = pool.ID
}

func ListPools() {
	log.Println(">>> List pools")
	pools, err := listPools(cli, authToken)
	handleError(err)
	printResponse(pools)
}

func GetPool() {
	log.Println(">>> Get pool")
	pool, err := getPool(cli, authToken, poolID)
	handleError(err)
	printResponse(pool)
}

func DeletePool() {
	log.Println(">>> Delete pool")
	err := deletePool(cli, authToken, poolID)
	handleError(err)
	log.Printf("pool %s deleted", poolID)
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

	//////////////////
	// organizations //
	//////////////////
	CreateOrg()
	ListOrgs()
	UpdateOrg()
	GetOrg()

	CreateOrgPool()
	ListOrgPools()
	GetOrgPool()
	UpdateOrgPool()

	///////////////
	// instances //
	///////////////
	WaitRepoInstance()
	ListRepoInstances()

	WaitOrgInstance()
	ListOrgInstances()

	ListInstances()
	GetInstance()

	///////////////
	// pools //
	///////////////
	CreatePool()
	ListPools()
	GetPool()

	/////////////
	// Cleanup //
	/////////////
	DisableRepoPool()
	DisableOrgPool()

	DeleteInstance(repoInstanceName)
	DeleteInstance(orgInstanceName)

	WaitRepoPoolNoInstances()
	WaitOrgPoolNoInstances()

	DeleteRepoPool()
	DeleteOrgPool()
	DeletePool()

	DeleteRepo()
	DeleteOrg()
}
