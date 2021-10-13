package main

import (
	"os"

	"github.com/pkg/errors"

	"github.com/smartcontractkit/chainlink/core/cmd"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/chainlink/core/sessions"
	"github.com/smartcontractkit/chainlink/core/store/config"
)

func main() {
	Run(NewProductionClient(), os.Args...)
	// Spam()
}

// Run runs the CLI, providing further command instructions by default.
func Run(client *cmd.Client, args ...string) {
	app := cmd.NewApp(client)
	client.Logger().WarnIf(app.Run(args), "Error running app")
}

// NewProductionClient configures an instance of the CLI to be used
// in production.
func NewProductionClient() *cmd.Client {
	cfg := config.NewGeneralConfig()

	prompter := cmd.NewTerminalPrompter()
	cookieAuth := cmd.NewSessionCookieAuthenticator(cfg, cmd.DiskCookieStore{Config: cfg})
	sr := sessions.SessionRequest{}
	sessionRequestBuilder := cmd.NewFileSessionRequestBuilder()
	if credentialsFile := cfg.AdminCredentialsFile(); credentialsFile != "" {
		var err error
		sr, err = sessionRequestBuilder.Build(credentialsFile)
		if err != nil && errors.Cause(err) != cmd.ErrNoCredentialFile && !os.IsNotExist(err) {
			logger.ProductionLogger(cfg).Fatalw("Error loading API credentials", "error", err, "credentialsFile", credentialsFile)
		}
	}
	return &cmd.Client{
		Renderer:                       cmd.RendererTable{Writer: os.Stdout},
		Config:                         cfg,
		AppFactory:                     cmd.ChainlinkAppFactory{},
		KeyStoreAuthenticator:          cmd.TerminalKeyStoreAuthenticator{Prompter: prompter},
		FallbackAPIInitializer:         cmd.NewPromptingAPIInitializer(prompter),
		Runner:                         cmd.ChainlinkRunner{},
		HTTP:                           cmd.NewAuthenticatedHTTPClient(cfg, cookieAuth, sr),
		CookieAuthenticator:            cookieAuth,
		FileSessionRequestBuilder:      sessionRequestBuilder,
		PromptingSessionRequestBuilder: cmd.NewPromptingSessionRequestBuilder(prompter),
		ChangePasswordPrompter:         cmd.NewChangePasswordPrompter(),
		PasswordPrompter:               cmd.NewPasswordPrompter(),
	}
}

// func Spam() {
//     s := `
// type                = "directrequest"
// schemaVersion       = 1
// name                = "example eth request event spec"
// contractAddress     = "0x613a38AC1659769640aaE063C651F48E0250454C"
// observationSource   = """
//     decode_log   [type=ethabidecodelog
//                  abi="OracleRequest(bytes32 indexed specId, address requester, bytes32 requestId, uint256 payment, address callbackAddr, bytes4 callbackFunctionId, uint256 cancelExpiration, uint256 dataVersion, bytes data)"
//                  data="$(jobRun.logData)"
//                  topics="$(jobRun.logTopics)"]
//     encode_tx  [type=ethabiencode
//                 abi="fulfill(bytes32 _requestId, uint256 _data)"
//                 data=<{
//                   "_requestId": $(decode_log.requestId),
//                   "_data": $(parse)
//                  }>
//                ]
//     fetch  [type=bridge name="test" requestData="{}"];
//     parse  [type=jsonparse path="foo"]
//     submit [type=ethtx to="$(decode_log.requester)" data="$(encode_tx)"]
//     decode_log -> fetch -> parse -> encode_tx -> submit
// """

// `
// }
