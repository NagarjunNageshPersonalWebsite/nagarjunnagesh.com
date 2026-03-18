// Moved inline scripts from index.html

// Tailwind config
tailwind.config = {
    theme: {
        extend: {
            colors: {
                yc: {
                    orange: '#F26522',
                    dark: '#1A1A1A',
                    light: '#F8F9FA',
                    gray: '#6B7280',
                    border: '#E5E7EB'
                }
            },
            fontFamily: {
                serif: ['Playfair Display', 'serif'],
                sans: ['Inter', 'sans-serif'],
            }
        }
    }
};

// Mobile menu behavior
document.addEventListener('DOMContentLoaded', () => {
    const btn = document.getElementById('mobile-menu-btn');
    const menu = document.getElementById('mobile-menu');
    if (!btn || !menu) return;
    const icon = btn.querySelector('i');

    btn.addEventListener('click', () => {
        menu.classList.toggle('hidden');
        if (menu.classList.contains('hidden')) {
            icon.classList.remove('fa-xmark');
            icon.classList.add('fa-bars');
        } else {
            icon.classList.remove('fa-bars');
            icon.classList.add('fa-xmark');
        }
    });

    // Close mobile menu when clicking a link
    const mobileLinks = menu.querySelectorAll('a');
    mobileLinks.forEach(link => {
        link.addEventListener('click', () => {
            menu.classList.add('hidden');
            icon.classList.remove('fa-xmark');
            icon.classList.add('fa-bars');
        });
    });
});

// Bottom nav active state (mobile)
document.addEventListener('DOMContentLoaded', () => {
    const sections = document.querySelectorAll('section[id]');
    const navLinks = document.querySelectorAll('nav.mobile-bottom-nav a');
    if (!sections.length || !navLinks.length) return;

    const setActive = () => {
        let current = '';
        sections.forEach(section => {
            const top = section.offsetTop;
            if (pageYOffset >= (top - (window.innerHeight * 0.4))) {
                current = section.id;
            }
        });

        navLinks.forEach(link => {
            link.classList.remove('text-yc-orange');
            link.classList.remove('active');
            if (link.getAttribute('href') === `#${current}`) {
                link.classList.add('text-yc-orange');
                link.classList.add('active');
            }
        });
    };

    window.addEventListener('scroll', setActive, { passive: true });
    setActive();
});
