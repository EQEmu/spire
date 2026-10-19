package eqemuserver

import (
	"errors"
	"fmt"
	"github.com/EQEmu/spire/internal/clientfiles"
	"github.com/EQEmu/spire/internal/database"
	"github.com/EQEmu/spire/internal/eqemuserverconfig"
	"github.com/EQEmu/spire/internal/http/routes"
	"github.com/EQEmu/spire/internal/pathmgmt"
	"github.com/EQEmu/spire/internal/spire"
	"github.com/labstack/echo/v4"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type PublicController struct {
	db             *database.Resolver
	eqemuserverapi *Client
	pathmgmt       *pathmgmt.PathManagement
	settings       *spire.Settings
	serverconfig   *eqemuserverconfig.Config
	updater        *Updater
	exporter       *clientfiles.Exporter
}

func NewPublicController(
	db *database.Resolver,
	api *Client,
	serverconfig *eqemuserverconfig.Config,
	pathmgmt *pathmgmt.PathManagement,
	settings *spire.Settings,
	updater *Updater,
	exporter *clientfiles.Exporter,
) *PublicController {
	return &PublicController{
		db:             db,
		eqemuserverapi: api,
		serverconfig:   serverconfig,
		pathmgmt:       pathmgmt,
		updater:        updater,
		settings:       settings,
		exporter:       exporter,
	}
}

func (a *PublicController) Routes() []*routes.Route {
	return []*routes.Route{
		routes.RegisterRoute(http.MethodGet, "eqemuserver/export-client-file/:type", a.exportClientFile, nil),
	}
}

type ExportType struct {
	arg  string
	file string
}

func (a *PublicController) exportClientFile(c echo.Context) error {
	// ?client=tob builds TOB client files with Spire's own exporter
	if c.QueryParam("client") == "tob" {
		return a.exportTobClientFile(c)
	}

	bin := a.pathmgmt.GetExportClientFilesBinPath()
	if _, err := os.Stat(bin); errors.Is(err, os.ErrNotExist) {
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{"error": fmt.Sprintf("Failed to find [%v] binary", a.pathmgmt.GetExportClientFilesBinPath())},
		)
	}

	t := []ExportType{
		{arg: "spells", file: "spells_us.txt"},
		{arg: "skills", file: "SkillCaps.txt"},
		{arg: "basedata", file: "BaseData.txt"},
		{arg: "dbstring", file: "dbstr_us.txt"},
	}

	exportType := c.Param("type")
	foundExport := false
	var export ExportType
	for _, e := range t {
		if e.arg == exportType {
			foundExport = true
			export = e
		}
	}

	if !foundExport {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Invalid export type"})
	}

	err := a.pathmgmt.MakeExportDirIfNotExists()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	cmd := exec.Command(bin, exportType)
	cmd.Dir = a.pathmgmt.GetEQEmuServerPath()
	_, err = cmd.Output()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	c.Response().Header().Add("Access-Control-Expose-Headers", "Content-Disposition")

	downloadFile := filepath.Join(a.pathmgmt.GetExportDir(), export.file)
	if _, err := os.Stat(downloadFile); !errors.Is(err, os.ErrNotExist) {
		if exportType == "spells" {
			return a.serveRoF2Spells(c, downloadFile)
		}

		return c.Attachment(downloadFile, filepath.Base(downloadFile))
	}

	return nil
}

// TOB client files, see clientfiles/export_tob.go
func (a *PublicController) exportTobClientFile(c echo.Context) error {
	var (
		file     string
		contents string
		err      error
	)

	db := a.db.GetEqemuDb()
	switch c.Param("type") {
	case "spells":
		file = "spells_us.txt"
		contents, err = a.exporter.ExportTobSpells(db)
	case "spellstrings":
		file = "spells_us_str.txt"
		contents, err = a.exporter.ExportTobSpellStrings(db)
	case "skills":
		file = "SkillCaps.txt"
		contents, err = a.exporter.ExportTobSkillCaps(db)
	case "basedata":
		file = "BaseData.txt"
		contents, err = a.exporter.ExportTobBaseData(db)
	case "dbstring":
		file = "dbstr_us.txt"
		contents, err = a.exporter.ExportTobDbStr(db)
	default:
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "Invalid export type"})
	}

	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	c.Response().Header().Add("Access-Control-Expose-Headers", "Content-Disposition")
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", file))

	return c.Blob(http.StatusOK, "text/plain", []byte(contents))
}

// serveRoF2Spells serves spells_us.txt without the spells the RoF2 client can't load
func (a *PublicController) serveRoF2Spells(c echo.Context, file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	lines := strings.Split(string(b), "\n")
	kept := lines[:0]
	for _, l := range lines {
		id, err := strconv.Atoi(strings.SplitN(l, "^", 2)[0])
		if err == nil && id > clientfiles.RoF2MaxSpellId {
			continue
		}

		kept = append(kept, l)
	}

	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%q", filepath.Base(file)))

	return c.Blob(http.StatusOK, "text/plain", []byte(strings.Join(kept, "\n")))
}
