const STORAGE_KEYS = {
  STATUS_FILTERS: "site-availability-status-filters",
  LABEL_FILTERS: "site-availability-label-filters",
  GROUP_BY_LABEL: "site-availability-group-by-label",
  SORT_ORDER: "site-availability-sort-order",
  SIDEBAR_COLLAPSED: "site-availability-sidebar-collapsed",
};

const storage = {
  save: (key, data) => {
    try {
      localStorage.setItem(key, JSON.stringify(data));
    } catch (error) {
      console.error("Error saving to localStorage:", error);
    }
  },

  load: (key, defaultValue = null) => {
    try {
      const item = localStorage.getItem(key);
      return item ? JSON.parse(item) : defaultValue;
    } catch (error) {
      console.error("Error loading from localStorage:", error);
      return defaultValue;
    }
  },

  remove: (key) => {
    try {
      localStorage.removeItem(key);
    } catch (error) {
      console.error("Error removing from localStorage:", error);
    }
  },
};

export const userPreferences = {
  saveStatusFilters: (filters) => {
    storage.save(STORAGE_KEYS.STATUS_FILTERS, filters);
  },

  loadStatusFilters: () => {
    return storage.load(STORAGE_KEYS.STATUS_FILTERS, []);
  },

  saveLabelFilters: (filters) => {
    storage.save(STORAGE_KEYS.LABEL_FILTERS, filters);
  },

  loadLabelFilters: () => {
    return storage.load(STORAGE_KEYS.LABEL_FILTERS, []);
  },

  saveGroupByLabel: (label) => {
    storage.save(STORAGE_KEYS.GROUP_BY_LABEL, label);
  },

  loadGroupByLabel: () => {
    return storage.load(STORAGE_KEYS.GROUP_BY_LABEL, null);
  },

  saveSortOrder: (order) => {
    storage.save(STORAGE_KEYS.SORT_ORDER, order);
  },

  loadSortOrder: () => {
    return storage.load(STORAGE_KEYS.SORT_ORDER, "name-asc");
  },

  saveSidebarCollapsed: (collapsed) => {
    storage.save(STORAGE_KEYS.SIDEBAR_COLLAPSED, collapsed);
  },

  loadSidebarCollapsed: () => {
    return storage.load(STORAGE_KEYS.SIDEBAR_COLLAPSED, false);
  },

  clearAll: () => {
    Object.values(STORAGE_KEYS).forEach((key) => {
      storage.remove(key);
    });
  },
};

export default userPreferences;
