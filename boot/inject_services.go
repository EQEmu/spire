package boot

import (
	"github.com/EQEmu/spire/internal/assets"
	"github.com/EQEmu/spire/internal/auditlog"
	"github.com/EQEmu/spire/internal/backup"
	"github.com/EQEmu/spire/internal/clientfiles"
	"github.com/EQEmu/spire/internal/connection"
	"github.com/EQEmu/spire/internal/desktop"
	"github.com/EQEmu/spire/internal/eqemuanalytics"
	"github.com/EQEmu/spire/internal/eqemuchangelog"
	"github.com/EQEmu/spire/internal/eqemuloginserver"
	"github.com/EQEmu/spire/internal/eqemuserver"
	"github.com/EQEmu/spire/internal/eqemuserverconfig"
	"github.com/EQEmu/spire/internal/github"
	"github.com/EQEmu/spire/internal/influx"
	"github.com/EQEmu/spire/internal/pathmgmt"
	"github.com/EQEmu/spire/internal/permissions"
	"github.com/EQEmu/spire/internal/questapi"
	"github.com/EQEmu/spire/internal/spire"
	"github.com/EQEmu/spire/internal/telnet"
	"github.com/EQEmu/spire/internal/unzip"
	"github.com/EQEmu/spire/internal/user"
	"github.com/EQEmu/spire/internal/websocket"
	pluralize "github.com/gertd/go-pluralize"
	"github.com/google/wire"
)

var serviceSet = wire.NewSet(
	influx.NewClient,
	connection.NewCreate,
	connection.NewCheck,
	github.NewGithubSourceDownloader,
	questapi.NewParseService,
	questapi.NewExamplesGithubSourcer,
	desktop.NewWebBoot,
	clientfiles.NewExporter,
	clientfiles.NewImporter,
	eqemuserverconfig.NewConfig,
	eqemuloginserver.NewConfig,
	pathmgmt.NewPathManagement,
	permissions.NewService,
	pluralize.NewClient,
	auditlog.NewUserEvent,
	assets.NewSpireAssets,
	eqemuchangelog.NewChangelog,
	eqemuanalytics.NewReleases,
	user.NewUser,
	spire.NewSettings,
	spire.NewInit,
	telnet.NewClient,
	eqemuserver.NewClient,
	backup.NewMysql,
	websocket.NewHandler,
	eqemuserver.NewUpdater,
	eqemuserver.NewLauncher,
	eqemuserver.NewQuestHotReloadWatcher,
	unzip.NewUnzipper,
	websocket.NewClientManager,
	eqemuserver.NewCrashLogWatcher,
)
