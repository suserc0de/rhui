// Initialize Lucide Icons
document.addEventListener("DOMContentLoaded", () => {
  if (window.lucide) {
    window.lucide.createIcons();
  }
});

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

// Mobile sidebar drawer toggle
function toggleMobileNav() {
  const drawer = document.getElementById("mobile-drawer");
  if (drawer) {
    drawer.classList.toggle("hidden");
  }
}
