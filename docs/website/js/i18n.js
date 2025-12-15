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
        'hero.desc': '复盘镜在后台自动记录你一天的工作痕迹，生成日报/周报与复盘总结。',
        'hero.desc2': '减少手写记录与回忆成本；每条结论都能点开回溯到来源证据；默认本地、可离线。',
        'hero.downloadBtn': '下载 Windows 版（Releases）',
        'hero.quickstart': '3 分钟上手',
        // Mockup
        'mockup.brand': '复盘镜',
        'mockup.dashboard': '仪表盘',
        'mockup.sessions': '片段流',
        'mockup.skills': '技能树',
        'mockup.reports': '报告',
        'mockup.diagnostics': '系统诊断',
        'mockup.systemHealth': '系统健康',
        'mockup.heartbeat': '心跳',
        'mockup.settings': '设置',
        'mockup.dashboardTitle': '仪表盘',
        'mockup.searchEvidence': '⌘K 搜索证据...',
        'mockup.todaySessions': '今日工作片段',
        'mockup.thirtyDayTotal': '/ 30天共 0',
        'mockup.thirtyDayChanges': '30天代码变更',
        'mockup.times': '次',
        'mockup.evidenceCoverage': '证据覆盖率',
        'mockup.medium': '中等',
        'mockup.weakEvidenceSessions': '个弱证据片段',
        'mockup.sessions2': '片段',
        'mockup.withDiff': '有Diff',
        'mockup.withBrowser': '有浏览',
        'mockup.focusDistribution': '专注分布',
        'mockup.coding': 'Coding',
        'mockup.reading': 'Reading',
        'mockup.meeting': 'Meeting',
        'mockup.dailySummaryReady': '每日总结已生成',
        'mockup.viewTodaySummary': '查看今日自动生成的工作回顾',
        'mockup.activityHeatmap': '活动热力图',
        'mockup.last30Days': '最近 30 天',
        'mockup.thirtyDaysAgo': '30天前',
        'mockup.today': '今天',
        // Features
        'features.title1': '把 ',
        'features.title2': '日报与复盘',
        'features.title3': ' 写得有据可查',
        'feature1.title': '自动日报/周报与复盘',
        'feature1.desc': '不靠记忆拼凑"今天做了什么"。系统会把一天拆成多个工作片段并生成摘要，也支持按周/周期聚合回顾，帮助你快速回看投入、成果与变化。',
        'feature2.title': '每条结论都能回溯',
        'feature2.desc': '不是"凭感觉"的总结。你可以点开看到每个工作片段的证据覆盖率与明细（代码变更、窗口行为、浏览历史等），把复盘变得可验证。',
        'feature3.title': '默认本地，离线也能用',
        'feature3.desc': '所有数据默认写入本地 SQLite；本地服务仅监听 127.0.0.1。不配 AI Key 也能生成规则版摘要；配置 AI 后再获得更强的语义化总结与建议。',
        // Tech specs
        'tech.title': '为 ',
        'tech.titleSuffix': ' 工程师打造',
        'tech1': '<strong>系统托盘常驻:</strong> 极低资源占用，静默运行于后台。',
        'tech2': '<strong>Git Diff 采集:</strong> 监听你配置的项目目录，捕获保存到磁盘的变更（无需 Commit）。',
        'tech3': '<strong>本地 API + SSE:</strong> 内置本地 HTTP Server（随机端口），UI 与状态页实时刷新，便于二次集成。',
        'tech4': '<strong>拒绝沉默失败:</strong> 内置状态页与诊断包导出，空态/异常会告诉你缺口在哪一层，以及下一步怎么修。',
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
        'hero.desc': 'WorkMirror runs in the background, recording your work traces and generating daily/weekly summaries.',
        'hero.desc2': 'No more manual logging; every conclusion links back to evidence; local-first and offline-ready.',
        'hero.downloadBtn': 'Download for Windows (Releases)',
        'hero.quickstart': 'Get Started in 3 min',
        // Mockup
        'mockup.brand': 'WorkMirror',
        'mockup.dashboard': 'Dashboard',
        'mockup.sessions': 'Sessions',
        'mockup.skills': 'Skills',
        'mockup.reports': 'Reports',
        'mockup.diagnostics': 'Diagnostics',
        'mockup.systemHealth': 'System Health',
        'mockup.heartbeat': 'Heartbeat',
        'mockup.settings': 'Settings',
        'mockup.dashboardTitle': 'Dashboard',
        'mockup.searchEvidence': '⌘K Search evidence...',
        'mockup.todaySessions': "Today's Sessions",
        'mockup.thirtyDayTotal': '/ 30d total 0',
        'mockup.thirtyDayChanges': '30-day code changes',
        'mockup.times': 'times',
        'mockup.evidenceCoverage': 'Evidence Coverage',
        'mockup.medium': 'Medium',
        'mockup.weakEvidenceSessions': 'weak evidence sessions',
        'mockup.sessions2': 'sessions',
        'mockup.withDiff': 'w/ Diff',
        'mockup.withBrowser': 'w/ Browser',
        'mockup.focusDistribution': 'Focus Distribution',
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
        'features.title3': ' with traceable evidence',
        'feature1.title': 'Auto Daily/Weekly Reports',
        'feature1.desc': 'No more piecing together "what did I do today" from memory. The system segments your day into work sessions with summaries, supporting weekly aggregation for quick review of effort, outcomes, and changes.',
        'feature2.title': 'Every Conclusion is Traceable',
        'feature2.desc': 'Not a "gut-feeling" summary. Click to see evidence coverage and details for each session (code changes, window behavior, browsing history, etc.), making reviews verifiable.',
        'feature3.title': 'Local-first, Works Offline',
        'feature3.desc': 'All data stored in local SQLite; local service listens only on 127.0.0.1. Rule-based summaries work without AI Key; configure AI for enhanced semantic summaries and suggestions.',
        // Tech specs
        'tech.title': 'Built for ',
        'tech.titleSuffix': ' Engineers',
        'tech1': '<strong>System Tray Resident:</strong> Minimal resource usage, runs silently in background.',
        'tech2': '<strong>Git Diff Collection:</strong> Monitors your configured project directories, captures disk-saved changes (no commit needed).',
        'tech3': '<strong>Local API + SSE:</strong> Built-in local HTTP Server (random port), real-time UI and status updates, easy integration.',
        'tech4': '<strong>No Silent Failures:</strong> Built-in status page and diagnostic export; empty states/errors tell you where the gap is and how to fix it.',
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
