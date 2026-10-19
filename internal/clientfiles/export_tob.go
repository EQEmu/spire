package clientfiles

import (
	"bufio"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// TOB client files
//
// The TOB spell file has 166 fields. Fields 0-164 map to these spells_new columns ("" = new in TOB, written as 0) and
// field 165 holds all the effects. Spell messages moved to spells_us_str.txt, and icon, memicon, player_1, Activated
// and a few unused fields were dropped.

const (
	tobEffectsField = 165
	tobZoneType     = 33
	tobEffectSlots  = 12
	tobEmptySpa     = "254"

	tobDbStrFactionName             = 45
	tobDbStrTributeName             = 48
	tobDbStrTributeDescription      = 49
	tobDbStrGuildTributeName        = 50
	tobDbStrGuildTributeDescription = 51
)

// range and rank are reserved words, hence the backticks
var tobSpellColumns = [tobEffectsField]string{
	"id", "name", "", "teleport_zone", "`range`", "aoerange", "pushback", "pushup", "cast_time",
	"recovery_time", "recast_time", "buffdurationformula", "buffduration", "AEDuration", "mana", "components1",
	"components2", "components3", "components4", "component_counts1", "component_counts2", "component_counts3",
	"component_counts4", "NoexpendReagent1", "NoexpendReagent2", "NoexpendReagent3", "NoexpendReagent4",
	"LightType", "goodEffect", "resisttype", "targettype", "basediff", "skill", "zonetype", "EnvironmentType",
	"TimeOfDay", "classes1", "classes2", "classes3", "classes4", "classes5", "classes6", "classes7", "classes8",
	"classes9", "classes10", "classes11", "classes12", "classes13", "classes14", "classes15", "classes16",
	"CastingAnim", "TargetAnim", "TravelType", "SpellAffectIndex", "disallow_sit", "deities0", "deities1",
	"deities2", "deities3", "deities4", "deities5", "deities6", "deities7", "deities8", "deities9", "deities10",
	"deities11", "deities12", "deities13", "deities14", "deities15", "deities16", "field142", "new_icon",
	"spellanim", "uninterruptable", "ResistDiff", "dot_stacking_exempt", "deleteable", "RecourseLink",
	"no_partial_resist", "field152", "short_buff_box", "descnum", "typedescnum", "effectdescnum",
	"effectdescnum2", "npc_no_los", "field160", "reflectable", "bonushate", "field163", "field164", "ldon_trap",
	"EndurCost", "EndurTimerIndex", "IsDiscipline", "HateAdded", "EndurUpkeep", "numhitstype", "numhits",
	"pvpresistbase", "pvpresistcalc", "pvpresistcap", "pvp_duration", "pvp_duration_cap", "pcnpc_only_flag",
	"cast_not_standing", "can_mgb", "nodispell", "npc_category", "npc_usefulness", "MinResist", "MaxResist",
	"viral_targets", "viral_timer", "nimbuseffect", "ConeStartAngle", "ConeStopAngle", "sneaking",
	"not_extendable", "field198", "field199", "suspendable", "viral_range", "songcap", "field203", "",
	"no_block", "field206", "spellgroup", "`rank`", "field209", "", "CastRestriction", "allowrest",
	"InCombat", "OutofCombat", "field215", "field217", "aemaxtargets", "maxtargets", "field220", "field221",
	"field222", "field223", "persistdeath", "field225", "field226", "min_dist", "min_dist_mod", "max_dist",
	"max_dist_mod", "min_range", "field232", "field233", "field234", "field235", "field236", "", "",
	"", "spell_category",
}

// rows as strings, formatted by MySQL like the server's export (NULL becomes "")
//
// The queries use SQL_NO_CACHE: with the MariaDB 10.5 query cache on, a result cached from here was served to
// export_client_files for the same query text, which then failed. MySQL 8 accepts the hint and ignores it.
func tobQuery(db *gorm.DB, query string) ([][]string, error) {
	rows, err := db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	var result [][]string
	for rows.Next() {
		values := make([]sql.RawBytes, len(cols))
		scan := make([]any, len(cols))
		for i := range values {
			scan[i] = &values[i]
		}

		if err := rows.Scan(scan...); err != nil {
			return nil, err
		}

		row := make([]string, len(cols))
		for i, v := range values {
			row[i] = string(v)
		}

		result = append(result, row)
	}

	return result, rows.Err()
}

// the TOB client needs CRLF, with LF it misreads a spell that has no effects
func tobJoinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}

	return strings.Join(lines, "\r\n") + "\r\n"
}

func (e *Exporter) ExportTobSpells(db *gorm.DB) (string, error) {
	var selects []string
	for _, c := range tobSpellColumns {
		if c == "" {
			c = "0"
		}
		selects = append(selects, c)
	}

	for i := 1; i <= tobEffectSlots; i++ {
		selects = append(
			selects,
			fmt.Sprintf("effectid%v", i),
			fmt.Sprintf("effect_base_value%v", i),
			fmt.Sprintf("effect_limit_value%v", i),
			fmt.Sprintf("formula%v", i),
			fmt.Sprintf("max%v", i),
		)
	}

	rows, err := tobQuery(db, fmt.Sprintf("SELECT SQL_NO_CACHE %v FROM spells_new ORDER BY id ASC", strings.Join(selects, ", ")))
	if err != nil {
		return "", err
	}

	var lines []string
	for _, row := range rows {
		fields := make([]string, tobEffectsField, tobEffectsField+1)
		copy(fields, row[:tobEffectsField])

		// TOB has no "any zone" -1
		if fields[tobZoneType] == "-1" {
			fields[tobZoneType] = "0"
		}

		// effects are slot|spa|base|base2|formula|max joined by $, empty slots left out
		var effects []string
		for slot := 0; slot < tobEffectSlots; slot++ {
			c := tobEffectsField + slot*5
			if row[c] == "" || row[c] == tobEmptySpa {
				continue
			}

			effects = append(
				effects,
				fmt.Sprintf("%v|%v|%v|%v|%v|%v", slot+1, row[c], row[c+1], row[c+2], row[c+3], row[c+4]),
			)
		}

		lines = append(lines, strings.Join(append(fields, strings.Join(effects, "$")), "^"))
	}

	return tobJoinLines(lines), nil
}

func (e *Exporter) ExportTobSpellStrings(db *gorm.DB) (string, error) {
	rows, err := tobQuery(
		db,
		"SELECT SQL_NO_CACHE id, you_cast, other_casts, cast_on_you, cast_on_other, spell_fades FROM spells_new ORDER BY id ASC",
	)
	if err != nil {
		return "", err
	}

	lines := []string{"#SPELLINDEX^CASTERMETXT^CASTEROTHERTXT^CASTEDMETXT^CASTEDOTHERTXT^SPELLGONE^"}
	for _, row := range rows {
		lines = append(lines, strings.Join(row, "^")+"^")
	}

	return tobJoinLines(lines), nil
}

// the client reads a string starting with a quote as a quoted token, so quote it like the stock file does
func tobDbStrText(text string) string {
	if text != "" && (text[0] == '"' || text[0] == '\'') {
		return "'" + strings.ReplaceAll(text, "'", "`") + "'"
	}

	return text
}

// id and type of a dbstr_us.txt line ("id^type^text^0^")
func tobDbStrKey(line string) (int, int, bool) {
	parts := strings.SplitN(line, "^", 3)
	if len(parts) != 3 {
		return 0, 0, false
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, false
	}

	t, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, false
	}

	return id, t, true
}

func (e *Exporter) ExportTobDbStr(db *gorm.DB) (string, error) {
	type key struct{ id, t int }
	dbStrings := map[key]string{}

	add := func(id, t int, text string) {
		if text != "" {
			dbStrings[key{id, t}] = fmt.Sprintf("%v^%v^%v^0^", id, t, tobDbStrText(text))
		}
	}

	rows, err := tobQuery(db, "SELECT SQL_NO_CACHE id, type, value FROM db_str ORDER BY id, type ASC")
	if err != nil {
		return "", err
	}

	for _, row := range rows {
		id, err1 := strconv.Atoi(row[0])
		t, err2 := strconv.Atoi(row[1])
		if err1 != nil || err2 != nil {
			continue
		}

		add(id, t, row[2])
	}

	// TOB reads faction and tribute names from db strings, which db_str usually doesn't have, so fill them in from
	// their own tables without replacing anything db_str already has
	fill := func(id, t int, text string) {
		if _, ok := dbStrings[key{id, t}]; !ok {
			add(id, t, text)
		}
	}

	rows, err = tobQuery(db, "SELECT SQL_NO_CACHE id, name FROM faction_list ORDER BY id ASC")
	if err != nil {
		e.logger.Warn().Err(err).Msg("failed to read factions")
	} else {
		for _, row := range rows {
			if id, err := strconv.Atoi(row[0]); err == nil {
				fill(id, tobDbStrFactionName, row[1])
			}
		}
	}

	rows, err = tobQuery(db, "SELECT SQL_NO_CACHE id, isguild, name, descr FROM tributes ORDER BY id ASC")
	if err != nil {
		e.logger.Warn().Err(err).Msg("failed to read tributes")
	} else {
		for _, row := range rows {
			id, err := strconv.Atoi(row[0])
			if err != nil {
				continue
			}

			nameType, descrType := tobDbStrTributeName, tobDbStrTributeDescription
			if row[1] != "0" {
				nameType, descrType = tobDbStrGuildTributeName, tobDbStrGuildTributeDescription
			}

			descr := strings.ReplaceAll(row[3], "</p><p>", "<br>")
			descr = strings.ReplaceAll(strings.ReplaceAll(descr, "<p>", ""), "</p>", "")

			fill(id, nameType, row[2])
			fill(id, descrType, descr)
		}
	}

	// the database holds only part of the client's strings; without the rest the client loses AA and spell
	// descriptions and many of its own texts, so every stock string the database doesn't replace is kept
	var unkeyed []string
	if e.pathmgmt != nil {
		basePath := filepath.Join(e.pathmgmt.GetEQEmuServerPath(), "export", "tob", "base", "dbstr_us.txt")
		if file, err := os.Open(basePath); err == nil {
			scanner := bufio.NewScanner(file)
			scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
			for scanner.Scan() {
				line := strings.TrimRight(scanner.Text(), "\r")
				id, t, ok := tobDbStrKey(line)
				if !ok {
					if line != "" {
						unkeyed = append(unkeyed, line)
					}
					continue
				}
				if _, ok := dbStrings[key{id, t}]; !ok {
					dbStrings[key{id, t}] = line
				}
			}
			file.Close()
			if err := scanner.Err(); err != nil {
				e.logger.Warn().Err(err).Str("path", basePath).Msg("failed to read the stock TOB dbstr_us.txt")
			}
		} else {
			e.logger.Warn().Str("path", basePath).Msg("no stock TOB dbstr_us.txt, exporting database strings only")
		}
	}

	keys := make([]key, 0, len(dbStrings))
	for k := range dbStrings {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].id != keys[j].id {
			return keys[i].id < keys[j].id
		}
		return keys[i].t < keys[j].t
	})

	lines := make([]string, 0, len(unkeyed)+len(keys))
	lines = append(lines, unkeyed...)
	for _, k := range keys {
		lines = append(lines, dbStrings[k])
	}

	return tobJoinLines(lines), nil
}

// same rows as the server's SkillCaps.txt and BaseData.txt exports, the TOB client reads them from Resources
func (e *Exporter) ExportTobSkillCaps(db *gorm.DB) (string, error) {
	return tobConcatExport(
		db,
		"SELECT SQL_NO_CACHE CONCAT_WS('^', `class_id`, `skill_id`, `level`, `cap`, `class_`) FROM skill_caps ORDER BY `class_id`, `skill_id`, `level` ASC",
	)
}

func (e *Exporter) ExportTobBaseData(db *gorm.DB) (string, error) {
	return tobConcatExport(
		db,
		"SELECT SQL_NO_CACHE CONCAT_WS('^', `level`, `class`, `hp`, `mana`, `end`, `hp_regen`, `end_regen`, `hp_fac`, `mana_fac`, `end_fac`) FROM base_data ORDER BY `level`, `class` ASC",
	)
}

func tobConcatExport(db *gorm.DB, query string) (string, error) {
	rows, err := tobQuery(db, query)
	if err != nil {
		return "", err
	}

	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, row[0])
	}

	return tobJoinLines(lines), nil
}
