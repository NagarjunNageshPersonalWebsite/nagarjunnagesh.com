// Client-side form handling with reCAPTCHA v3 and basic validation
// Replace RECAPTCHA_SITE_KEY and API_URL placeholders before deploying

(() => {
  const RECAPTCHA_SITE_KEY = 'REPLACE_WITH_RECAPTCHA_SITE_KEY'; // replace
  const API_URL = 'REPLACE_WITH_YOUR_API_GATEWAY_URL'; // replace, e.g. https://api.example.com/send
  const MIN_SCORE = 0.5;

  function qs(sel, el = document) { return el.querySelector(sel); }

  async function timeoutFetch(resource, options = {}, ms = 10000) {
    const controller = new AbortController();
    const id = setTimeout(() => controller.abort(), ms);
    try {
      const res = await fetch(resource, { ...options, signal: controller.signal });
      clearTimeout(id);
      return res;
    } catch (err) {
      clearTimeout(id);
      throw err;
    }
  }

  function showFeedback(msg, srOnly = false) {
    const el = qs('#contact-feedback');
    if (!el) return;
    el.className = srOnly ? 'sr-only' : '';
    el.textContent = msg;
  }

  function validateEmail(email) {
    return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email);
  }

  document.addEventListener('DOMContentLoaded', () => {
    const form = qs('#contact-form');
    if (!form) return;

    form.addEventListener('submit', async (ev) => {
      ev.preventDefault();
      showFeedback('Sending...', false);

      const hp = form.querySelector('input[name="hp_website"]');
      if (hp && hp.value.trim() !== '') {
        // Honeypot triggered
        showFeedback('Bot detected (honeypot).', false);
        return;
      }

      const name = (form.name.value || '').trim();
      const email = (form.email.value || '').trim();
      const message = (form.message.value || '').trim();

      if (!name || !email || !message) {
        showFeedback('Please fill all required fields.', false);
        return;
      }
      if (!validateEmail(email)) {
        showFeedback('Please provide a valid email.', false);
        return;
      }
      if (message.length < 10 || message.length > 1000) {
        showFeedback('Message length must be between 10 and 1000 characters.', false);
        return;
      }

      try {
        // get reCAPTCHA token
        if (!window.grecaptcha || RECAPTCHA_SITE_KEY.includes('REPLACE')) {
          showFeedback('reCAPTCHA not configured. Contact admin.', false);
          return;
        }

        const token = await grecaptcha.execute(RECAPTCHA_SITE_KEY, { action: 'contact_form' });

        const payload = {
          name, email, message,
          recaptchaToken: token,
          source: location.origin
        };

        const res = await timeoutFetch(API_URL, {
          method: 'POST',
          mode: 'cors',
          credentials: 'omit',
          headers: {
            'Content-Type': 'application/json',
            'X-Requested-With': 'XMLHttpRequest'
          },
          body: JSON.stringify(payload)
        }, 12000);

        if (!res.ok) {
          const text = await res.text().catch(() => 'Server error');
          showFeedback('Error: ' + text, false);
          return;
        }

        showFeedback('Message sent — thank you!', false);
        form.reset();
      } catch (err) {
        if (err.name === 'AbortError') showFeedback('Request timed out. Try again later.', false);
        else showFeedback('Unexpected error. Try again later.', false);
      }
    });
  });
})();
