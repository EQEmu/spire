<template>
  <div>

    <eq-window title="EverQuest Client Asset Exports (Publicly Downloadable Links)">
      <div class="eq-alert text-center">
        <i class="fe fe-info"></i>
        Files are generated from your server's database on every download. Links work without login, so they can be
        shared with players. If you link from an SSL site, browsers block the download, so you will need to proxy the request
      </div>

      <div class="row mt-3">
        <div
          v-for="g in assetGroups"
          :key="g.title"
          class="col-lg-6 mb-3"
        >
          <div class="client-group p-3 h-100">
            <div class="text-center mb-3">
              <div class="eq-header" style="font-size: 22px">{{ g.title }}</div>
              <div class="client-group-subtitle">{{ g.subtitle }}</div>
            </div>

            <div
              v-for="r in g.assets"
              :key="(r.client || '') + r.path"
              class="client-asset d-flex align-items-center"
            >
              <a
                class="eq-button client-asset-button text-center"
                :href="download(r)"
                :target="(r.client || '') + r.path"
                :title="'Download ' + r.asset"
              >
                <i class="fa fa-download mr-1"></i> {{ r.asset }}
              </a>

              <div class="flex-grow-1 ml-3">
                <div class="font-weight-bold">{{ r.description }}</div>
                <div class="client-asset-destination">{{ r.destination }}</div>
              </div>

              <i
                class="fa fa-link client-asset-copy ml-3"
                title="Copy public link"
                @click="copyLink(r)"
              ></i>
            </div>
          </div>
        </div>
      </div>
    </eq-window>

  </div>
</template>

<script>
import util       from "util";
import {SpireApi} from "@/app/api/spire-api";
import EqWindow   from "@/components/eq-ui/EQWindow.vue";
import ClipBoard  from "@/app/clipboard/clipboard";
import {Notify}   from "@/app/Notify";

export default {
  components: { EqWindow },
  data() {
    return {
      assetGroups: [
        {
          title: "RoF2",
          subtitle: "Rain of Fear 2 and older clients",
          assets: [
            {asset: "spells_us.txt", path: "spells", description: "Spells", destination: "EverQuest folder"},
            {asset: "dbstr_us.txt", path: "dbstring", description: "Strings (AAs, spell descriptions)", destination: "EverQuest folder"},
            {asset: "SkillCaps.txt", path: "skills", description: "Skill caps", destination: "Resources folder"},
            {asset: "BaseData.txt", path: "basedata", description: "Base HP, mana and endurance", destination: "Resources folder"},
          ]
        },
        {
          title: "TOB",
          subtitle: "The Outer Brood client",
          assets: [
            {asset: "spells_us.txt", path: "spells", client: "tob", description: "Spells", destination: "EverQuest folder"},
            {asset: "spells_us_str.txt", path: "spellstrings", client: "tob", description: "Spell messages", destination: "EverQuest folder"},
            {asset: "dbstr_us.txt", path: "dbstring", client: "tob", description: "Strings (AAs, spell descriptions, currencies, tributes)", destination: "EverQuest folder"},
            {asset: "SkillCaps.txt", path: "skills", client: "tob", description: "Skill caps", destination: "Resources folder"},
            {asset: "BaseData.txt", path: "basedata", client: "tob", description: "Base HP, mana and endurance", destination: "Resources folder"},
          ]
        },
      ]
    }
  },
  methods: {
    download(a) {
      return util.format(
        "%s/eqemuserver/export-client-file/%s%s",
        SpireApi.getBaseV1Path(),
        a.path,
        a.client ? "?client=" + a.client : "",
      )
    },
    copyLink(a) {
      ClipBoard.copyFromText(this.download(a))
      Notify.toast("Copied link to clipboard!");
    },
  },
}
</script>

<style scoped>
.client-group {
  background-color: rgba(0, 0, 0, 0.25);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 6px;
}

.client-group-subtitle {
  opacity: 0.7;
  font-size: 13px;
}

.client-asset {
  padding: 10px 4px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.client-asset-button {
  min-width: 180px;
}

.client-asset-destination {
  opacity: 0.6;
  font-size: 12px;
}

.client-asset-copy {
  cursor: pointer;
  opacity: 0.6;
  font-size: 16px;
}

.client-asset-copy:hover {
  opacity: 1;
}
</style>
