// Moved inline scripts from index.html

// Mobile menu removed — using bottom navigation on small screens

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
