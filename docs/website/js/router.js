// WorkMirror Website - View Router

// Track current doc for language switching
let currentDocId = 'install';

// Cache for loaded documents
const docCache = {};

// Mobile menu toggle with accessibility support
function toggleMobileMenu() {
    try {
        const menu = document.getElementById('mobile-menu');
        const btn = document.getElementById('mobile-menu-btn');
        if (menu && btn) {
            const isHidden = menu.classList.contains('hidden');
            menu.classList.toggle('hidden');
            // Update ARIA attribute for accessibility
            btn.setAttribute('aria-expanded', isHidden ? 'true' : 'false');
        }
    } catch (error) {
        console.error('Error toggling mobile menu:', error);
    }
}

// Toggle docs submenu in mobile navigation
function toggleDocsSubmenu() {
    try {
        const submenu = document.getElementById('mobile-docs-submenu');
        const chevron = document.getElementById('mobile-docs-chevron');
        if (submenu && chevron) {
            const isHidden = submenu.classList.contains('hidden');
            submenu.classList.toggle('hidden');
            chevron.style.transform = isHidden ? 'rotate(180deg)' : '';
        }
    } catch (error) {
        console.error('Error toggling docs submenu:', error);
    }
}

// Toggle doc dropdown in docs view navbar
function toggleDocDropdown() {
    try {
        const menu = document.getElementById('doc-dropdown-menu');
        const chevron = document.getElementById('doc-dropdown-chevron');
        if (menu && chevron) {
            const isHidden = menu.classList.contains('hidden');
            menu.classList.toggle('hidden');
            chevron.style.transform = isHidden ? 'rotate(180deg)' : '';
        }
    } catch (error) {
        console.error('Error toggling doc dropdown:', error);
    }
}

// Show doc and close the dropdown
function showDocAndCloseDropdown(docId) {
    showDoc(docId);
    const menu = document.getElementById('doc-dropdown-menu');
    const chevron = document.getElementById('doc-dropdown-chevron');
    if (menu) menu.classList.add('hidden');
    if (chevron) chevron.style.transform = '';
}

// Close dropdown when clicking outside
document.addEventListener('click', function(event) {
    const dropdown = document.getElementById('doc-dropdown-menu');
    const btn = document.getElementById('doc-dropdown-btn');
    if (dropdown && btn && !dropdown.contains(event.target) && !btn.contains(event.target)) {
        dropdown.classList.add('hidden');
        const chevron = document.getElementById('doc-dropdown-chevron');
        if (chevron) chevron.style.transform = '';
    }
});

// Wrapper functions for mobile menu navigation
function handleNavAndCloseMobile(sectionId) {
    try {
        handleNavClick(sectionId);
        toggleMobileMenu();
    } catch (error) {
        console.error('Error navigating:', error);
    }
}

function switchViewAndCloseMobile(viewName, docId) {
    try {
        switchView(viewName, docId);
        toggleMobileMenu();
    } catch (error) {
        console.error('Error switching view:', error);
    }
}

// 更新 URL 参数（不刷新页面）
function updateUrlParams(params) {
    const url = new URL(window.location);
    Object.entries(params).forEach(([key, value]) => {
        if (value) {
            url.searchParams.set(key, value);
        } else {
            url.searchParams.delete(key);
        }
    });
    window.history.replaceState({}, '', url);
}

// Simple View Router
function switchView(viewName, docId = null) {
    // 1. Handle View Toggling
    const views = ['home', 'docs'];
    views.forEach(v => {
        const el = document.getElementById(`view-${v}`);
        if (v === viewName) {
            el.classList.remove('hidden');
            el.classList.add('block');
        } else {
            el.classList.add('hidden');
            el.classList.remove('block');
        }
    });

    // 2. Handle Scroll Reset
    window.scrollTo(0, 0);

    // 3. Update URL params
    if (viewName === 'home') {
        updateUrlParams({ doc: null });
    }

    // 4. If switching to docs, optional deep link
    if (viewName === 'docs' && docId) {
        showDoc(docId);
    }
}

// Handle Nav Clicks (Conditional Logic)
function handleNavClick(sectionId) {
    const homeView = document.getElementById('view-home');
    if (homeView.classList.contains('hidden')) {
        // If on docs, switch to home first
        switchView('home');
        // Wait for render then scroll
        setTimeout(() => {
            document.getElementById(sectionId).scrollIntoView({ behavior: 'smooth' });
        }, 10);
    } else {
        // Just scroll
        document.getElementById(sectionId).scrollIntoView({ behavior: 'smooth' });
    }
}

// Load document content from external file
async function loadDocContent(docId, lang) {
    const cacheKey = `${lang}/${docId}`;

    // Return cached content if available
    if (docCache[cacheKey]) {
        return docCache[cacheKey];
    }

    try {
        const response = await fetch(`docs/${lang}/${docId}.html`);
        if (!response.ok) {
            // Fallback to Chinese if English version doesn't exist
            if (lang === 'en') {
                return loadDocContent(docId, 'zh');
            }
            throw new Error(`Failed to load ${cacheKey}`);
        }
        const content = await response.text();
        docCache[cacheKey] = content;
        return content;
    } catch (error) {
        console.error('Error loading doc:', error);
        return `<article class="doc-content"><h1>加载失败</h1><p>无法加载文档内容。</p></article>`;
    }
}

// Internal Doc Router
async function showDoc(docId) {
    currentDocId = docId;

    // Update URL params
    updateUrlParams({ doc: docId });

    const container = document.getElementById('doc-container');
    const lang = typeof getCurrentLang === 'function' ? getCurrentLang() : 'zh';

    // Show loading state
    container.innerHTML = '<div class="text-zinc-500 text-center py-12">Loading...</div>';

    // Load and display content
    const content = await loadDocContent(docId, lang);
    container.innerHTML = content;

    // Update Sidebar Active State
    document.querySelectorAll('[id^="nav-"]').forEach(el => el.classList.remove('nav-active'));
    const navBtn = document.getElementById(`nav-${docId}`);
    if (navBtn) navBtn.classList.add('nav-active');

    // Update mobile/tablet nav title
    updateDocNavTitle(docId, lang);
}

// Update the doc navigation bar title
function updateDocNavTitle(docId, lang) {
    const titleEl = document.getElementById('current-doc-title');
    if (!titleEl) return;

    // Map docId to i18n key
    const docTitleMap = {
        'overview': 'docs.overview',
        'architecture': 'docs.architecture',
        'install': 'docs.install',
        'config': 'docs.config',
        'api': 'docs.api',
        'faq': 'docs.faq',
        'privacy': 'docs.privacyPolicy',
        'terms': 'docs.terms'
    };

    const i18nKey = docTitleMap[docId];
    if (i18nKey) {
        titleEl.setAttribute('data-i18n', i18nKey);
        // If i18n is loaded, apply translation
        if (typeof applyTranslations === 'function') {
            applyTranslations();
        }
    }
}

// Get current doc ID (for i18n module)
function getCurrentDocId() {
    return currentDocId;
}

// Clear doc cache (called when language changes)
function clearDocCache() {
    Object.keys(docCache).forEach(key => delete docCache[key]);
}

// 从 URL 参数恢复页面状态
function initFromUrl() {
    const urlParams = new URLSearchParams(window.location.search);
    const docId = urlParams.get('doc');

    // 有效的文档 ID 列表
    const validDocs = ['overview', 'architecture', 'install', 'config', 'api', 'faq', 'privacy', 'terms'];

    if (docId && validDocs.includes(docId)) {
        switchView('docs', docId);
    }
}

// 页面加载时初始化
document.addEventListener('DOMContentLoaded', initFromUrl);
