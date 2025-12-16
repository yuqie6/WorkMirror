// WorkMirror Website - Internationalization (i18n)

const translations = {
    zh: {
        // Nav
        'nav.brand': '复盘镜',
        'nav.features': '特性',
        'nav.privacy': '隐私',
        'nav.docs': '文档',
        'nav.download': '下载 Windows 版',
        // Hero
        'hero.tagline': 'Local-first & Privacy by default',
        'hero.title1': '每天自动复盘你的工作',
        'hero.title2': '一键生成日报',
        'hero.desc': '复盘镜在后台自动记录你一天的工作内容，生成日报/周报与复盘总结。',
        'hero.desc2': '减少手写记录与回忆成本；每条结论都能点开看到数据来源；默认本地、可离线。',
        'hero.downloadBtn': '下载 Windows 版（Releases）',
        'hero.quickstart': '3 分钟上手',
        // Mockup
        'mockup.brand': '复盘镜',
        'mockup.dashboard': '概览',
        'mockup.sessions': '工作记录',
        'mockup.skills': '技能成长',
        'mockup.reports': '报告',
        'mockup.diagnostics': '运行状态',
        'mockup.systemHealth': '系统状态',
        'mockup.heartbeat': '状态',
        'mockup.settings': '设置',
        'mockup.dashboardTitle': '概览',
        'mockup.searchEvidence': '⌘K 搜索记录...',
        'mockup.todaySessions': '今天的工作片段',
        'mockup.thirtyDayTotal': '/ 30天共 0',
        'mockup.thirtyDayChanges': '30天代码修改',
        'mockup.times': '次',
        'mockup.evidenceCoverage': '记录完整度',
        'mockup.medium': '部分',
        'mockup.weakEvidenceSessions': '条记录不完整',
        'mockup.sessions2': '条记录',
        'mockup.withDiff': '含代码',
        'mockup.withBrowser': '含网页',
        'mockup.focusDistribution': '时间分布',
        'mockup.coding': '编码',
        'mockup.reading': '阅读',
        'mockup.meeting': '会议',
        'mockup.dailySummaryReady': '今日总结已生成',
        'mockup.viewTodaySummary': '查看今天自动生成的工作回顾',
        'mockup.activityHeatmap': '活动热力图',
        'mockup.last30Days': '最近 30 天',
        'mockup.thirtyDaysAgo': '30天前',
        'mockup.today': '今天',
        // Features
        'features.title1': '把 ',
        'features.title2': '日报与复盘',
        'features.title3': ' 写得有据可查',
        'feature1.title': '自动日报/周报与复盘',
        'feature1.desc': '不靠记忆拼凑"今天做了什么"。系统会把一天拆成多个工作片段并生成摘要，也支持按周聚合回顾，帮助你快速回看投入、成果与变化。',
        'feature2.title': '每条结论都能回溯',
        'feature2.desc': '不是"凭感觉"的总结。你可以点开看到每条记录的完整度与明细（代码修改、应用使用、网页浏览等），把复盘变得可验证。',
        'feature3.title': '默认本地，离线也能用',
        'feature3.desc': '所有数据默认写入本地 SQLite；本地服务仅监听 127.0.0.1。不配 AI Key 也能生成规则版摘要；配置 AI 后再获得更强的语义化总结与建议。',
        // Tech specs
        'tech.title': '为 ',
        'tech.titleSuffix': ' 工程师打造',
        'tech1': '<strong>系统托盘常驻:</strong> 极低资源占用，静默运行于后台。',
        'tech2': '<strong>代码监控:</strong> 监听你配置的项目目录，捕获保存到磁盘的变更（无需 Commit）。',
        'tech3': '<strong>本地 API + 实时更新:</strong> 内置本地 HTTP 服务（随机端口），UI 实时刷新，便于二次集成。',
        'tech4': '<strong>遇到问题不迷茫:</strong> 内置状态页与诊断包导出，空白页/异常会告诉你问题在哪里，以及下一步怎么修。',
        // Docs sidebar
        'docs.userGuide': '使用指南',
        'docs.overview': '了解复盘镜',
        'docs.architecture': '工作原理（高级）',
        'docs.install': '快速开始',
        'docs.config': '设置与隐私',
        'docs.api': '本地接口（高级）',
        'docs.faq': '常见问题',
        'docs.legal': 'Legal',
        'docs.privacyPolicy': '隐私策略',
        'docs.terms': '服务条款',
        'docs.backHome': '返回首页',
        // Footer
        'footer.brand': '复盘镜',
        'footer.copyright': '© 2025 复盘镜（WorkMirror）. Built for Engineers.',
        'footer.docs': '文档',
        'footer.privacy': '隐私策略',
    },
    en: {
        // Nav
        'nav.brand': 'WorkMirror',
        'nav.features': 'Features',
        'nav.privacy': 'Privacy',
        'nav.docs': 'Docs',
        'nav.download': 'Download for Windows',
        // Hero
        'hero.tagline': 'Local-first & Privacy by default',
        'hero.title1': 'Auto-review your daily work',
        'hero.title2': 'One-click daily reports',
        'hero.desc': 'WorkMirror runs in the background, recording your work activities and generating daily/weekly summaries.',
        'hero.desc2': 'No more manual logging; every conclusion links back to source data; local-first and offline-ready.',
        'hero.downloadBtn': 'Download for Windows (Releases)',
        'hero.quickstart': 'Get Started in 3 min',
        // Mockup
        'mockup.brand': 'WorkMirror',
        'mockup.dashboard': 'Overview',
        'mockup.sessions': 'Work Records',
        'mockup.skills': 'Skills',
        'mockup.reports': 'Reports',
        'mockup.diagnostics': 'Status',
        'mockup.systemHealth': 'System Status',
        'mockup.heartbeat': 'Status',
        'mockup.settings': 'Settings',
        'mockup.dashboardTitle': 'Overview',
        'mockup.searchEvidence': '⌘K Search records...',
        'mockup.todaySessions': "Today's Work Sessions",
        'mockup.thirtyDayTotal': '/ 30d total 0',
        'mockup.thirtyDayChanges': '30-day code changes',
        'mockup.times': 'times',
        'mockup.evidenceCoverage': 'Record Completeness',
        'mockup.medium': 'Partial',
        'mockup.weakEvidenceSessions': 'incomplete records',
        'mockup.sessions2': 'records',
        'mockup.withDiff': 'w/ Code',
        'mockup.withBrowser': 'w/ Browser',
        'mockup.focusDistribution': 'Time Distribution',
        'mockup.coding': 'Coding',
        'mockup.reading': 'Reading',
        'mockup.meeting': 'Meeting',
        'mockup.dailySummaryReady': 'Daily Summary Ready',
        'mockup.viewTodaySummary': "View today's auto-generated work review",
        'mockup.activityHeatmap': 'Activity Heatmap',
        'mockup.last30Days': 'Last 30 days',
        'mockup.thirtyDaysAgo': '30 days ago',
        'mockup.today': 'Today',
        // Features
        'features.title1': 'Write ',
        'features.title2': 'daily reports',
        'features.title3': ' with traceable sources',
        'feature1.title': 'Auto Daily/Weekly Reports',
        'feature1.desc': 'No more piecing together "what did I do today" from memory. The system segments your day into work sessions with summaries, supporting weekly aggregation for quick review of effort, outcomes, and changes.',
        'feature2.title': 'Every Conclusion is Traceable',
        'feature2.desc': 'Not a "gut-feeling" summary. Click to see record completeness and details for each session (code changes, app usage, web browsing, etc.), making reviews verifiable.',
        'feature3.title': 'Local-first, Works Offline',
        'feature3.desc': 'All data stored in local SQLite; local service listens only on 127.0.0.1. Rule-based summaries work without AI Key; configure AI for enhanced semantic summaries and suggestions.',
        // Tech specs
        'tech.title': 'Built for ',
        'tech.titleSuffix': ' Engineers',
        'tech1': '<strong>System Tray Resident:</strong> Minimal resource usage, runs silently in background.',
        'tech2': '<strong>Code Monitoring:</strong> Monitors your configured project directories, captures disk-saved changes (no commit needed).',
        'tech3': '<strong>Local API + Real-time Updates:</strong> Built-in local HTTP service (random port), real-time UI updates, easy integration.',
        'tech4': '<strong>Clear Error Messages:</strong> Built-in status page and diagnostic export; empty pages/errors tell you where the problem is and how to fix it.',
        // Docs sidebar
        'docs.userGuide': 'User Guide',
        'docs.overview': 'About WorkMirror',
        'docs.architecture': 'How It Works (Advanced)',
        'docs.install': 'Quick Start',
        'docs.config': 'Settings & Privacy',
        'docs.api': 'Local API (Advanced)',
        'docs.faq': 'FAQ',
        'docs.legal': 'Legal',
        'docs.privacyPolicy': 'Privacy Policy',
        'docs.terms': 'Terms of Service',
        'docs.backHome': 'Back to Home',
        // Footer
        'footer.brand': 'WorkMirror',
        'footer.copyright': '© 2025 WorkMirror. Built for Engineers.',
        'footer.docs': 'Documentation',
        'footer.privacy': 'Privacy Policy',
    }
};

// 优先从 URL 参数读取语言，其次 localStorage，默认中文
function getInitialLang() {
    const urlParams = new URLSearchParams(window.location.search);
    const urlLang = urlParams.get('lang');
    if (urlLang === 'en' || urlLang === 'zh') {
        return urlLang;
    }
    return localStorage.getItem('workmirror-site-lang') || 'zh';
}

let currentLang = getInitialLang();

// 更新 URL 参数（不刷新页面）
function updateUrlLang(lang) {
    const url = new URL(window.location);
    url.searchParams.set('lang', lang);
    window.history.replaceState({}, '', url);
}

function toggleLanguage() {
    currentLang = currentLang === 'zh' ? 'en' : 'zh';
    localStorage.setItem('workmirror-site-lang', currentLang);
    updateUrlLang(currentLang);
    applyTranslations();
    updateLangSwitch();
    // Clear doc cache and refresh current doc view if in docs view
    if (typeof clearDocCache === 'function') {
        clearDocCache();
    }
    if (typeof currentDocId !== 'undefined' && currentDocId && !document.getElementById('view-docs').classList.contains('hidden')) {
        showDoc(currentDocId);
    }
}

function updateLangSwitch() {
    const btn = document.getElementById('lang-switch');
    const btnMobile = document.getElementById('lang-switch-mobile');
    const text = currentLang === 'zh' ? 'EN' : '中';
    if (btn) btn.textContent = text;
    if (btnMobile) btnMobile.textContent = text;
}

function applyTranslations() {
    const t = translations[currentLang];
    document.querySelectorAll('[data-i18n]').forEach(el => {
        const key = el.getAttribute('data-i18n');
        if (t[key]) {
            // For elements with HTML content (like tech specs), use innerHTML
            if (key.startsWith('tech') && key !== 'tech.title' && key !== 'tech.titleSuffix') {
                el.innerHTML = t[key];
            } else {
                el.textContent = t[key];
            }
        }
    });
    // Update html lang attribute
    document.documentElement.lang = currentLang === 'zh' ? 'zh-CN' : 'en';
}

function getCurrentLang() {
    return currentLang;
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', () => {
    // 确保 URL 上有 lang 参数
    updateUrlLang(currentLang);
    updateLangSwitch();
    applyTranslations();
});
