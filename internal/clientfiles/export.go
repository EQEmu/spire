package clientfiles

import (
	"database/sql"
	"fmt"
	"github.com/EQEmu/spire/internal/database"
	"github.com/EQEmu/spire/internal/logger"
	"github.com/EQEmu/spire/internal/models"
	"github.com/EQEmu/spire/internal/pathmgmt"
	"gorm.io/gorm"
	"strings"
)

type Exporter struct {
	logger   *logger.AppLogger
	pathmgmt *pathmgmt.PathManagement
}

func NewExporter(logger *logger.AppLogger, pathmgmt *pathmgmt.PathManagement) *Exporter {
	return &Exporter{logger: logger, pathmgmt: pathmgmt}
}

func (e *Exporter) getDatabase(g *gorm.DB) *sql.DB {
	// get database instance
	db, err := g.DB()
	if err != nil {
		e.logger.Warn().Err(err).Msg("failed to get database instance")
	}

	return db
}

// RoF2MaxSpellId is the highest spell id the RoF2 client can load, a spells_us.txt with higher ids breaks it
const RoF2MaxSpellId = 44999

func (e *Exporter) ExportSpells(db *gorm.DB) string {
	var entries []map[string]interface{}
	db.Model(&models.SpellsNew{}).Where("id <= ?", RoF2MaxSpellId).Find(&entries)

	columns, err := database.GetTableSchema(db, "spells_new")
	if err != nil {
		e.logger.Warn().Err(err).Msg("failed to get table schema")
	}

	var rows []string
	for _, e := range entries {
		var cols []string
		for _, column := range columns {
			cols = append(cols, fmt.Sprintf("%v", e[column.Column]))
		}

		rows = append(rows, strings.Join(cols, "^"))
	}

	return strings.Join(rows, "\n")
}

func (e *Exporter) ExportDbStr(db *gorm.DB) string {
	var entries []map[string]interface{}
	db.Model(&models.DbStr{}).Find(&entries)

	columns, err := database.GetTableSchema(db, "db_str")
	if err != nil {
		e.logger.Warn().Err(err).Msg("failed to get table schema")
	}

	var rows []string
	for _, e := range entries {
		var cols []string
		for _, column := range columns {
			cols = append(cols, fmt.Sprintf("%v", e[column.Column]))
		}
		cols = append(cols, "0")

		rows = append(rows, strings.Join(cols, "^"))
	}

	return strings.Join(rows, "\n")
}
