/**
 * BlogPostsApp is a Vue.js component for managing blog posts.
 * It provides a table view with multi-filter tags, sorting, pagination,
 * and CRUD operations for posts.
 *
 * The filter system follows the statsstore pattern: each filter is a
 * {field, operator, value} condition. Multiple conditions are AND-combined.
 * Active filters show as removable badge tags. The URL encodes all
 * conditions as JSON so filter state is shareable/bookmarkable.
 */
const BlogPostsApp = {
  data() {
    return {
      // UI state
      loading: true,
      showCreateModal: false,
      showFilterModal: false,
      creating: false,

      // Feature flags (injected from the server via initScript)
      aiEnabled: (typeof aiEnabled !== 'undefined') ? aiEnabled : false,

      // Post data
      posts: [],
      totalPosts: 0,
      totalPages: 0,

      // Pagination (0-indexed)
      currentPage: 0,
      perPage: 10,

      // Applied conditions (sent to the server)
      conditions: [],

      // Modal state — cloned from conditions when opening
      modalConditions: [],

      // Sorting
      sortByColumn: 'created_at',
      sortOrder: 'desc',

      // Create post form
      createForm: {
        title: ''
      },

      // Filter condition options — must match the Go CondField* constants.
      conditionOptions: [
        { value: 'search',    label: 'Search',         operators: ['contains'], inputType: 'text',  placeholder: 'title or content...' },
        { value: 'status',    label: 'Status',         operators: ['equals'],   inputType: 'select', placeholder: '' },
        { value: 'slug',      label: 'Slug',           operators: ['equals'],   inputType: 'text',  placeholder: 'e.g. my-first-post' },
        { value: 'date_from', label: 'Created after',  operators: ['equals'],   inputType: 'date',  placeholder: '' },
        { value: 'date_to',   label: 'Created before', operators: ['equals'],   inputType: 'date',  placeholder: '' },
      ],

      // Operator display labels
      opLabels: { equals: '=', contains: 'contains' },

      // Status dropdown options
      statusOptions: [
        { value: 'draft',       label: 'Draft' },
        { value: 'published',   label: 'Published' },
        { value: 'unpublished', label: 'Unpublished' },
        { value: 'trash',       label: 'Trash' },
      ]
    };
  },

  computed: {
    visiblePages() {
      const pages = [];
      const start = Math.max(0, this.currentPage - 2);
      const end = Math.min(this.totalPages - 1, this.currentPage + 2);
      for (let i = start; i <= end; i++) {
        pages.push(i);
      }
      return pages;
    }
  },

  mounted() {
    this.loadFromURL();
    this.loadPosts();
  },

  methods: {
    // === Filter modal ===

    openFilterModal() {
      this.modalConditions = this.conditions.map(c => ({ ...c }));
      if (this.modalConditions.length === 0) {
        this.modalConditions.push({ field: '', operator: 'equals', value: '' });
      }
      this.showFilterModal = true;
    },

    closeFilterModal() {
      this.showFilterModal = false;
    },

    addModalCondition() {
      this.modalConditions.push({ field: '', operator: 'equals', value: '' });
    },

    removeModalCondition(i) {
      this.modalConditions.splice(i, 1);
    },

    clearModalConditions() {
      this.modalConditions = [];
    },

    onFieldChange(c) {
      const opt = this.getFieldOpt(c.field);
      if (opt) {
        c.operator = opt.operators[0];
      }
      c.value = '';
    },

    applyConditions() {
      this.conditions = this.modalConditions.filter(c => c.field && c.value.trim());
      this.showFilterModal = false;
      this.currentPage = 0;
      this.loadPosts();
      this.updateURL();
    },

    clearConditions() {
      this.conditions = [];
      this.currentPage = 0;
      this.loadPosts();
      this.updateURL();
    },

    removeCondition(i) {
      this.conditions.splice(i, 1);
      this.currentPage = 0;
      this.loadPosts();
      this.updateURL();
    },

    // === Data loading ===

    async loadPosts() {
      this.loading = true;
      try {
        const response = await fetch(urlPostsLoad, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({
            page: this.currentPage,
            per_page: this.perPage,
            sort_by: this.sortByColumn,
            sort_order: this.sortOrder,
            conditions: this.conditions,
          }),
        });
        const data = await response.json();
        if (data.status === 'success') {
          this.posts = data.data?.posts || [];
          this.totalPosts = data.data?.total || 0;
          this.totalPages = data.data?.total_pages || 0;
        } else {
          Swal.fire({
            icon: 'error',
            title: 'Error',
            text: data.message || 'Failed to load posts'
          });
        }
      } catch (error) {
        console.error('Error loading posts:', error);
        Swal.fire({
          icon: 'error',
          title: 'Error',
          text: error.message || 'Failed to load posts'
        });
      } finally {
        this.loading = false;
      }
    },

    // === Sorting ===

    sortBy(column) {
      if (this.sortByColumn === column) {
        this.sortOrder = this.sortOrder === 'asc' ? 'desc' : 'asc';
      } else {
        this.sortByColumn = column;
        this.sortOrder = 'asc';
      }
      this.currentPage = 0;
      this.loadPosts();
      this.updateURL();
    },

    // === Pagination ===

    goToPage(page) {
      if (page < 0 || page >= this.totalPages) return;
      this.currentPage = page;
      this.loadPosts();
      this.updateURL();
    },

    // === URL encode/decode ===

    updateURL() {
      const params = new URLSearchParams();
      // Preserve controller param
      const existing = new URLSearchParams(window.location.search);
      const controller = existing.get('controller');
      if (controller) params.set('controller', controller);

      if (this.conditions.length > 0) {
        params.set('filters', JSON.stringify(this.conditions));
      }
      if (this.currentPage > 0) {
        params.set('page', String(this.currentPage));
      }
      if (this.perPage !== 10) {
        params.set('per_page', String(this.perPage));
      }
      if (this.sortByColumn !== 'created_at') {
        params.set('sort_by', this.sortByColumn);
      }
      if (this.sortOrder !== 'desc') {
        params.set('sort_order', this.sortOrder);
      }

      const qs = params.toString();
      const newURL = qs ? window.location.pathname + '?' + qs : window.location.pathname;
      history.replaceState(null, '', newURL);
    },

    loadFromURL() {
      const params = new URLSearchParams(window.location.search);

      // Try JSON filters first (new format)
      const filtersRaw = params.get('filters');
      if (filtersRaw) {
        try {
          const parsed = JSON.parse(filtersRaw);
          if (Array.isArray(parsed)) {
            this.conditions = parsed.filter(c => c && c.field && c.value);
          }
        } catch (e) { /* ignore malformed */ }
      } else {
        // Fallback: read legacy individual params
        if (params.get('search')) this.conditions.push({ field: 'search', operator: 'contains', value: params.get('search') });
        if (params.get('status')) this.conditions.push({ field: 'status', operator: 'equals', value: params.get('status') });
        if (params.get('slug')) this.conditions.push({ field: 'slug', operator: 'equals', value: params.get('slug') });
        if (params.get('date_from')) this.conditions.push({ field: 'date_from', operator: 'equals', value: params.get('date_from') });
        if (params.get('date_to')) this.conditions.push({ field: 'date_to', operator: 'equals', value: params.get('date_to') });
      }

      const pageRaw = params.get('page');
      if (pageRaw) this.currentPage = parseInt(pageRaw, 10) || 0;

      const perPageRaw = params.get('per_page');
      if (perPageRaw) this.perPage = parseInt(perPageRaw, 10) || 10;

      this.sortByColumn = params.get('sort_by') || 'created_at';
      this.sortOrder = params.get('sort_order') || 'desc';
    },

    // === CRUD ===

    async deletePost(post) {
      const result = await Swal.fire({
        icon: 'warning',
        title: 'Delete Post?',
        text: `Are you sure you want to delete "${post.title}"?`,
        showCancelButton: true,
        confirmButtonText: 'Delete',
        cancelButtonText: 'Cancel',
        confirmButtonColor: '#dc3545'
      });
      if (!result.isConfirmed) return;

      try {
        const response = await fetch(urlPostDelete, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ post_id: post.id })
        });
        const data = await response.json();
        if (data.status === 'success') {
          Swal.fire({ icon: 'success', title: 'Deleted', text: 'Post deleted successfully', timer: 1500, showConfirmButton: false });
          this.loadPosts();
        } else {
          Swal.fire({ icon: 'error', title: 'Error', text: data.message || 'Failed to delete post' });
        }
      } catch (error) {
        Swal.fire({ icon: 'error', title: 'Error', text: 'Failed to delete post' });
      }
    },

    async createPost() {
      if (!this.createForm.title) return;
      this.creating = true;
      try {
        const response = await fetch(urlPostCreate, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ title: this.createForm.title })
        });
        const data = await response.json();
        if (data.status === 'success') {
          Swal.fire({ icon: 'success', title: 'Success', text: 'Post created successfully', timer: 1500, showConfirmButton: false });
          this.closeCreateModal();
          window.open(urlPostUpdate.replace('POST_ID_PLACEHOLDER', data.data.id), '_blank');
          this.loadPosts();
        } else {
          Swal.fire({ icon: 'error', title: 'Error', text: data.message || 'Failed to create post' });
        }
      } catch (error) {
        Swal.fire({ icon: 'error', title: 'Error', text: 'Failed to create post' });
      } finally {
        this.creating = false;
      }
    },

    // === Helpers ===

    openCreateModal() {
      this.createForm.title = '';
      this.showCreateModal = true;
    },

    closeCreateModal() {
      this.showCreateModal = false;
      this.createForm.title = '';
    },

    formatDate(dateString) {
      if (!dateString) return '-';
      const date = new Date(dateString);
      return date.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' });
    },

    getWebsitePostUrl(postId, slug) {
      return `/blog/post/${postId}/${slug}`;
    },

    getAiPostContentUrl(postId) {
      return urlAiPostContentUpdate.replace('POST_ID_PLACEHOLDER', postId);
    },

    getPostUpdateUrl(postId) {
      return urlPostUpdate.replace('POST_ID_PLACEHOLDER', postId);
    },

    // === Filter option helpers ===

    getFieldOpt(value) {
      return this.conditionOptions.find(o => o.value === value);
    },
    fieldLabel(v) { const o = this.getFieldOpt(v); return o ? o.label : v; },
    opLabel(op) { return this.opLabels[op] || op; },
    getFieldOperators(v) { const o = this.getFieldOpt(v); return o ? o.operators : ['equals']; },
    getFieldInputType(v) { const o = this.getFieldOpt(v); return o ? o.inputType : 'text'; },
    getFieldPlaceholder(v) { const o = this.getFieldOpt(v); return o ? o.placeholder : ''; },
  }
};

// Mount the app when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
  loadVueIfNeeded((err) => {
    if (err) { console.error('Vue load failed:', err); return; }
    const { createApp } = Vue;
    const el = document.getElementById('blog-posts-app');
    if (el) {
      createApp(BlogPostsApp).mount('#blog-posts-app');
    }
  });
});
