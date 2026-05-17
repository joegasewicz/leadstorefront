import "./styles.scss";
import Alpine from "alpinejs";
import Sortable from "sortablejs";
import { AlignCenter, AlignLeft, AlignRight, createIcons } from "lucide";

document.documentElement.classList.add("js-enabled");

const attributionParameterNames = [
  "fbclid",
  "gclid",
  "msclkid",
  "ttclid",
  "li_fat_id",
  "twclid",
  "epik",
  "rdt_cid",
  "utm_source",
  "utm_medium",
  "utm_campaign",
  "utm_content",
  "utm_term",
  "affid",
  "aid",
  "partner_id",
  "partner",
  "subid",
  "sub_id",
  "clickid",
  "click_ref",
  "campaign_id",
  "ref",
  "source",
  "cid",
  "market",
  "locale",
  "currency",
  "destination",
  "hotel_id",
  "offer_id",
  "checkin",
  "checkout",
  "adults",
  "children"
];

function persistAttributionLocally() {
  const params = new URLSearchParams(window.location.search);
  const captured: Record<string, string> = {};
  attributionParameterNames.forEach((name) => {
    const value = params.get(name)?.trim();
    if (value) {
      captured[name] = value;
    }
  });
  if (Object.keys(captured).length === 0) {
    return;
  }
  const now = new Date().toISOString();
  const existingRaw = window.localStorage.getItem("leadstorefront_attribution");
  let existing: { first_touch?: Record<string, string>; latest_touch?: Record<string, string> } = {};
  if (existingRaw) {
    try {
      existing = JSON.parse(existingRaw);
    } catch (_error) {
      existing = {};
    }
  }
  window.localStorage.setItem("leadstorefront_attribution", JSON.stringify({
    first_touch: existing.first_touch || captured,
    latest_touch: captured,
    captured_at: now,
    landing_path: `${window.location.pathname}${window.location.search}`
  }));
}

persistAttributionLocally();

type StorefrontThemeColors = {
  primary: string;
  accent: string;
  background: string;
  text: string;
  surface: string;
};

type StorefrontThemeContentColumn = {
  heading: string;
  body: string;
};

type StorefrontThemeSectionOptions = {
  content_kind?: string;
  title?: string;
  description?: string;
  columns?: StorefrontThemeContentColumn[];
};

type StorefrontThemeSection = {
  id: string;
  name: string;
  type: string;
  enabled: boolean;
  container_style?: string;
  text_alignments?: Record<string, string>;
  options: StorefrontThemeSectionOptions;
};

type StorefrontThemeConfig = {
  colors: StorefrontThemeColors;
  sections: StorefrontThemeSection[];
};

declare global {
  interface Window {
    Alpine: typeof Alpine;
  }
}

const defaultTheme: StorefrontThemeConfig = {
  colors: {
    primary: "#67e8f9",
    accent: "#38bdf8",
    background: "#020617",
    text: "#ffffff",
    surface: "#0f172a"
  },
  sections: [
    { id: "hero", name: "Hero", type: "hero", enabled: true, options: {} },
    { id: "lead-form", name: "Lead form", type: "content", enabled: true, options: { content_kind: "lead_form" } },
    { id: "about", name: "About", type: "content", enabled: true, options: { content_kind: "about" } },
    { id: "products", name: "Products", type: "content", enabled: true, options: { content_kind: "products" } },
    { id: "articles", name: "Articles", type: "content", enabled: true, options: { content_kind: "articles" } },
    { id: "footer", name: "Footer", type: "footer", enabled: true, options: {} }
  ]
};

function sectionID(name: string, type: string): string {
  const base = (name || type || "section")
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
  return `${base || "section"}-${Date.now().toString(36)}`;
}

function normalizeSection(section: Partial<StorefrontThemeSection>): StorefrontThemeSection {
  const type = section.type || "content";
  const options = { ...(section.options || {}) };
  if (type === "content") {
    options.content_kind = options.content_kind || "custom";
    if (options.content_kind === "custom") {
      options.title = options.title || section.name || "Content";
      options.description = options.description || "";
      options.columns = (options.columns || []).map((column) => ({
        heading: column.heading || "",
        body: column.body || ""
      }));
    } else {
      delete options.title;
      delete options.description;
      delete options.columns;
    }
  }
  return {
    id: section.id || sectionID(section.name || type, type),
    name: section.name || type || "Section",
    type,
    enabled: section.enabled ?? true,
    container_style: section.container_style || "",
    text_alignments: normalizeTextAlignments(section.text_alignments || {}),
    options: type === "content" ? options : {}
  };
}

function normalizeTextAlignments(alignments: Record<string, string>): Record<string, string> {
  const normalized: Record<string, string> = {};
  ["h1", "h2", "h3", "h4", "h5", "h6", "p"].forEach((element) => {
    if (["left", "center", "right"].includes(alignments[element])) {
      normalized[element] = alignments[element];
    }
  });
  return normalized;
}

Alpine.data("storefrontThemeEditor", () => ({
  colors: { ...defaultTheme.colors },
  sections: defaultTheme.sections.map((section) => normalizeSection(section)),
  selectedSectionID: "hero",
  newSectionName: "",
  newSectionType: "content",
  newContentKind: "custom",
  alignmentTargets: [
    { key: "h1", label: "H1" },
    { key: "h2", label: "H2" },
    { key: "h3", label: "H3" },
    { key: "h4", label: "H4" },
    { key: "h5", label: "H5" },
    { key: "h6", label: "H6" },
    { key: "p", label: "P" }
  ],
  alignmentOptions: [
    { value: "left", icon: "align-left", label: "Align left" },
    { value: "center", icon: "align-center", label: "Align center" },
    { value: "right", icon: "align-right", label: "Align right" }
  ],
  sortable: null as Sortable | null,

  init() {
    const input = this.$refs.designConfig as HTMLInputElement | undefined;
    if (input?.value) {
      try {
        const config = JSON.parse(input.value) as Partial<StorefrontThemeConfig>;
        this.colors = { ...defaultTheme.colors, ...(config.colors || {}) };
        if (config.sections?.length) {
          this.sections = config.sections.map((section) => normalizeSection(section));
        }
      } catch (_error) {
        this.colors = { ...defaultTheme.colors };
      }
    }
    this.selectedSectionID = this.sections[0]?.id || "";

    this.$watch("colors", () => this.updateDesignConfig());
    this.$watch("sections", () => this.updateDesignConfig());

    this.$nextTick(() => {
      this.bindSortable();
      this.updateDesignConfig();
      this.refreshIcons();
    });
  },

  bindSortable() {
    const sectionList = this.$refs.sections as HTMLElement | undefined;
    if (!sectionList) {
      return;
    }
    this.sortable?.destroy();
    this.sortable = Sortable.create(sectionList, {
      animation: 150,
      handle: ".theme-section-handle",
      onEnd: () => {
        const orderedIDs = Array.from(sectionList.querySelectorAll<HTMLElement>("[data-section-id]")).map((element) => element.dataset.sectionId || "");
        this.sections = orderedIDs
          .map((id) => this.sections.find((section) => section.id === id))
          .filter((section): section is StorefrontThemeSection => Boolean(section));
        if (!this.selectedSection()) {
          this.selectedSectionID = this.sections[0]?.id || "";
        }
        this.updateDesignConfig();
      }
    });
  },

  selectedSection() {
    return this.sections.find((section) => section.id === this.selectedSectionID) || this.sections[0] || null;
  },

  selectSection(id: string) {
    this.selectedSectionID = id;
    this.refreshIcons();
  },

  refreshIcons() {
    this.$nextTick(() => createIcons({ icons: { AlignLeft, AlignCenter, AlignRight } }));
  },

  textAlignment(section: StorefrontThemeSection, element: string) {
    section.text_alignments = section.text_alignments || {};
    return section.text_alignments[element] || "";
  },

  setTextAlignment(section: StorefrontThemeSection, element: string, alignment: string) {
    section.text_alignments = section.text_alignments || {};
    if (section.text_alignments[element] === alignment) {
      delete section.text_alignments[element];
    } else {
      section.text_alignments[element] = alignment;
    }
    this.updateDesignConfig();
  },

  setSectionType(section: StorefrontThemeSection, type: string) {
    section.type = type || "content";
    if (section.type !== "content") {
      section.options = {};
    } else {
      section.options = { content_kind: section.options.content_kind || "custom" };
      this.ensureCustomContent(section);
    }
    this.updateDesignConfig();
  },

  setContentKind(section: StorefrontThemeSection, contentKind: string) {
    section.options = {
      ...section.options,
      content_kind: contentKind || "custom"
    };
    this.ensureCustomContent(section);
    this.updateDesignConfig();
  },

  ensureCustomContent(section: StorefrontThemeSection) {
    if (section.type !== "content" || section.options.content_kind !== "custom") {
      return;
    }
    section.options.title = section.options.title || section.name || "Content";
    section.options.description = section.options.description || "";
    section.options.columns = section.options.columns || [];
  },

  addColumn(section: StorefrontThemeSection) {
    this.ensureCustomContent(section);
    section.options.columns?.push({ heading: "", body: "" });
    this.updateDesignConfig();
  },

  removeColumn(section: StorefrontThemeSection, index: number) {
    section.options.columns = (section.options.columns || []).filter((_column, columnIndex) => columnIndex !== index);
    this.updateDesignConfig();
  },

  addSection() {
    const type = this.newSectionType || "content";
    const name = this.newSectionName.trim() || (type === "hero" ? "Hero" : type === "footer" ? "Footer" : "Content");
    const section = normalizeSection({
      id: sectionID(name, type),
      name,
      type,
      enabled: true,
      container_style: "",
      text_alignments: {},
      options: type === "content" ? { content_kind: this.newContentKind || "custom", title: name, description: "", columns: [] } : {}
    });
    this.sections.push(section);
    this.selectedSectionID = section.id;
    this.newSectionName = "";
    this.newSectionType = "content";
    this.newContentKind = "custom";
    this.$nextTick(() => {
      this.bindSortable();
      this.updateDesignConfig();
      this.refreshIcons();
    });
  },

  removeSection(id: string) {
    if (this.sections.length <= 1) {
      return;
    }
    this.sections = this.sections.filter((section) => section.id !== id);
    if (this.selectedSectionID === id) {
      this.selectedSectionID = this.sections[0]?.id || "";
    }
    this.$nextTick(() => this.updateDesignConfig());
  },

  updateDesignConfig() {
    const input = this.$refs.designConfig as HTMLInputElement | undefined;
    if (!input) {
      return;
    }
    input.value = JSON.stringify({
      colors: this.colors,
      sections: this.sections.map((section) => ({
        id: section.id,
        name: section.name,
        type: section.type,
        enabled: section.enabled,
        container_style: section.container_style || "",
        text_alignments: normalizeTextAlignments(section.text_alignments || {}),
        options: section.type === "content" ? {
          content_kind: section.options.content_kind || "custom",
          title: section.options.content_kind === "custom" ? section.options.title || section.name : undefined,
          description: section.options.content_kind === "custom" ? section.options.description || "" : undefined,
          columns: section.options.content_kind === "custom" ? section.options.columns || [] : undefined
        } : {}
      }))
    });
  }
}));

window.Alpine = Alpine;
Alpine.start();
