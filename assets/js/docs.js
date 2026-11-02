// Navigation Data Source (Single Source of Truth)
const DOCS_NAV_ITEMS = [
  {
    title: "Getting Started",
    path: "getting-started",
    icon: "play-circle"
  },
  {
    title: "Components & Slots",
    path: "components",
    icon: "box"
  },
  {
    title: "Syntax & Directives",
    path: "syntax",
    icon: "code"
  },
  {
    title: "Benchmarks",
    path: "benchmarks",
    icon: "gauge"
  },
  {
    title: "SaaS Showcase",
    path: "showcase",
    icon: "layout-dashboard"
  },
  {
    title: "Production Validation",
    path: "validation",
    icon: "shield-check"
  },
  {
    title: "Realtime SSE",
    path: "sse",
    icon: "radio"
  },
  {
    title: "Framework Adapters",
    path: "frameworks",
    icon: "layers"
  },
  {
    title: "SEO & Sitemap",
    path: "seo",
    icon: "search"
  }
];

// Helper to determine root-relative prefix based on current pathname
function getBasePrefix() {
  const path = window.location.pathname;
  // If we are at root or /index.html or /rhui/
  if (path.endsWith("/rhui/") || path.endsWith("/rhui") || path === "/" || path.endsWith("/index.html") && !path.includes("/getting-started/") && !path.includes("/components/") && !path.includes("/syntax/") && !path.includes("/benchmarks/") && !path.includes("/showcase/") && !path.includes("/validation/") && !path.includes("/sse/") && !path.includes("/frameworks/") && !path.includes("/seo/")) {
    return "./";
  }
  return "../";
}

// Render dynamic sidebar and header navigation across all pages
function renderDynamicNavigation() {
  const currentPath = window.location.pathname.replace(/\/index\.html$/, "").replace(/\/$/, "");
  const basePrefix = getBasePrefix();

  // 1. Render Desktop Sidebar if #sidebarNav exists
  const sidebarNav = document.getElementById("sidebarNav");
  if (sidebarNav) {
    sidebarNav.innerHTML = DOCS_NAV_ITEMS.map(item => {
      const targetUrl = `${basePrefix}${item.path}/`;
      const isActive = currentPath.endsWith(`/${item.path}`);

      if (isActive) {
        return `
          <a href="${targetUrl}" class="flex items-center px-3 py-2 rounded-lg bg-brand-500/10 text-brand-400 border border-brand-500/20 font-semibold transition-all">
            <i data-lucide="${item.icon}" class="w-4 h-4 mr-2.5"></i>
            <span>${item.title}</span>
          </a>
        `;
      }
      return `
        <a href="${targetUrl}" class="flex items-center px-3 py-2 rounded-lg text-slate-400 hover:bg-slate-800/60 hover:text-white transition-colors">
          <i data-lucide="${item.icon}" class="w-4 h-4 mr-2.5"></i>
          <span>${item.title}</span>
        </a>
      `;
    }).join("");
  }

  // 2. Render Desktop Header Nav if #desktopHeaderNav exists
  const desktopHeaderNav = document.getElementById("desktopHeaderNav");
  if (desktopHeaderNav) {
    desktopHeaderNav.innerHTML = DOCS_NAV_ITEMS.map(item => {
      const targetUrl = `${basePrefix}${item.path}/`;
      const isActive = currentPath.endsWith(`/${item.path}`);

      if (isActive) {
        return `<a href="${targetUrl}" class="text-brand-400 hover:text-brand-300 font-semibold transition-colors">${item.title}</a>`;
      }
      return `<a href="${targetUrl}" class="text-slate-400 hover:text-white transition-colors">${item.title}</a>`;
    }).join("");
  }

  // 3. Render Mobile Drawer Nav if #mobileDrawerNav exists
  const mobileDrawerNav = document.getElementById("mobileDrawerNav");
  if (mobileDrawerNav) {
    mobileDrawerNav.innerHTML = DOCS_NAV_ITEMS.map(item => {
      const targetUrl = `${basePrefix}${item.path}/`;
      const isActive = currentPath.endsWith(`/${item.path}`);

      if (isActive) {
        return `<a href="${targetUrl}" class="text-brand-400 font-semibold py-2.5 border-b border-slate-800 flex items-center"><i data-lucide="${item.icon}" class="w-4 h-4 mr-2.5"></i>${item.title}</a>`;
      }
      return `<a href="${targetUrl}" class="text-slate-300 hover:text-brand-400 py-2.5 border-b border-slate-800 flex items-center"><i data-lucide="${item.icon}" class="w-4 h-4 mr-2.5"></i>${item.title}</a>`;
    }).join("");
  }

  // Refresh Lucide Icons after DOM injection
  if (window.lucide) {
    window.lucide.createIcons();
  }
}

// Copy code block to clipboard
function copyCode(btn) {
  const container = btn.closest('.code-container');
  const codeElem = container ? container.querySelector('code') : null;
  if (!codeElem) return;

  const text = codeElem.innerText;
  navigator.clipboard.writeText(text).then(() => {
    const originalHTML = btn.innerHTML;
    btn.innerHTML = `<span class="text-xs text-emerald-400 font-sans font-semibold flex items-center space-x-1">
      <svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>
      <span>Copied!</span>
    </span>`;
    setTimeout(() => {
      btn.innerHTML = originalHTML;
      if (window.lucide) window.lucide.createIcons();
    }, 2000);
  });
}

// Fallback mobile drawer toggle helper
function toggleMobileNav() {
  const drawer = document.getElementById("mobileDrawer");
  if (drawer) {
    drawer.classList.toggle("hidden");
  }
}

// Initialize on DOM load
document.addEventListener("DOMContentLoaded", () => {
  renderDynamicNavigation();

  // Bind mobile menu button
  const mobileBtn = document.getElementById("mobileMenuBtn");
  const mobileDrawer = document.getElementById("mobileDrawer");
  if (mobileBtn && mobileDrawer) {
    mobileBtn.addEventListener("click", () => {
      mobileDrawer.classList.toggle("hidden");
    });
  }
});
