// Vega AI Landing Page - JavaScript Functionality
// Main JavaScript file handling interactions and animations

// Tailwind CSS Configuration
tailwind.config = {
  theme: {
    fontFamily: {
      'sans': ['DM Sans', 'Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
      'heading': ['Space Grotesk', 'DM Sans', 'ui-sans-serif', 'system-ui', 'sans-serif']
    },
    extend: {
      colors: {
        primary: '#0D9488',
        'primary-dark': '#0B7A70',
        secondary: '#F59E0B',
      },
      animation: {
        'float': 'float 6s ease-in-out infinite',
        'float-delayed': 'float 6s ease-in-out 2s infinite',
        'twinkle': 'twinkle 3s ease-in-out infinite',
        'pulse-slow': 'pulse 4s cubic-bezier(0.4, 0, 0.6, 1) infinite',
        'slide-up': 'slideUp 0.5s ease-out',
        'slide-down': 'slideDown 0.3s ease-out',
        'glow': 'glow 2s ease-in-out infinite',
        'orbit': 'orbit 20s linear infinite',
        'reverse-orbit': 'reverseOrbit 30s linear infinite',
      },
      keyframes: {
        float: {
          '0%, 100%': { transform: 'translateY(0px)' },
          '50%': { transform: 'translateY(-20px)' },
        },
        twinkle: {
          '0%, 100%': { opacity: 0.2 },
          '50%': { opacity: 1 },
        },
        slideUp: {
          '0%': { transform: 'translateY(20px)', opacity: 0 },
          '100%': { transform: 'translateY(0)', opacity: 1 },
        },
        slideDown: {
          '0%': { transform: 'translateY(-10px)', opacity: 0 },
          '100%': { transform: 'translateY(0)', opacity: 1 },
        },
        glow: {
          '0%, 100%': { opacity: 0.5, transform: 'scale(1)' },
          '50%': { opacity: 1, transform: 'scale(1.05)' },
        },
        orbit: {
          '0%': { transform: 'rotate(0deg) translateX(100px) rotate(0deg)' },
          '100%': { transform: 'rotate(360deg) translateX(100px) rotate(-360deg)' },
        },
        reverseOrbit: {
          '0%': { transform: 'rotate(0deg) translateX(120px) rotate(0deg)' },
          '100%': { transform: 'rotate(-360deg) translateX(120px) rotate(360deg)' },
        }
      },
      backgroundImage: {
        'gradient-radial': 'radial-gradient(var(--tw-gradient-stops))',
      }
    }
  }
}

// Global mobile menu toggle function
function toggleMobileMenu() {
  const mobileMenu = document.getElementById('mobile-menu');
  const isActive = mobileMenu.classList.contains('active');
  
  mobileMenu.classList.toggle('active');
  document.body.style.overflow = !isActive ? 'hidden' : '';

  // Prevent scrolling on iOS
  if (!isActive) {
    document.body.style.position = 'fixed';
    document.body.style.width = '100%';
    document.body.style.top = `-${window.scrollY}px`;
  } else {
    const scrollY = document.body.style.top;
    document.body.style.position = '';
    document.body.style.width = '';
    document.body.style.top = '';
    window.scrollTo(0, parseInt(scrollY || '0') * -1);
  }
}

// DOM Ready Handler - Initialize all functionality when page loads
document.addEventListener('DOMContentLoaded', function() {

  // Mobile menu functionality
  const mobileMenuButton = document.getElementById('mobile-menu-button');
  const mobileMenuClose = document.getElementById('mobile-menu-close');
  const mobileMenu = document.getElementById('mobile-menu');
  
  if (mobileMenuButton && mobileMenuClose && mobileMenu) {
    const mobileMenuLinks = mobileMenu.querySelectorAll('a[href^="#"]');
    
    mobileMenuButton.addEventListener('click', toggleMobileMenu);
    mobileMenuClose.addEventListener('click', toggleMobileMenu);

    // Close mobile menu when clicking on links
    mobileMenuLinks.forEach(link => {
      link.addEventListener('click', () => {
        toggleMobileMenu();
      });
    });
  }

  // Smooth scroll for navigation links
  document.querySelectorAll('a[href^="#"]').forEach(anchor => {
    anchor.addEventListener('click', function (e) {
      e.preventDefault();
      const target = document.querySelector(this.getAttribute('href'));
      if (target) {
        const offset = 80; // Account for fixed nav
        const targetPosition = target.getBoundingClientRect().top + window.pageYOffset - offset;
        window.scrollTo({
          top: targetPosition,
          behavior: 'smooth'
        });
      }
    });
  });

  // Enhanced scroll effects with navigation state
  const nav = document.querySelector('nav');
  if (nav) {
    let lastScroll = 0;

    function handleScroll() {
      const currentScroll = window.pageYOffset;

      // Add scrolled class for enhanced styling
      nav.classList.toggle('scrolled', currentScroll > 50);

      // Hide/show nav on scroll (mobile) with smoother animation
      if (window.innerWidth < 768) {
        const isScrollingDown = currentScroll > lastScroll && currentScroll > 100;
        nav.style.transform = isScrollingDown ? 'translateY(-100%)' : 'translateY(0)';
      }

      lastScroll = currentScroll;
    }

    window.addEventListener('scroll', handleScroll, { passive: true });
  }

  // Intersection Observer for fade-in animations
  const observerOptions = {
    threshold: 0.1,
    rootMargin: '0px 0px -50px 0px'
  };

  const observer = new IntersectionObserver(function(entries) {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        entry.target.classList.add('visible');
      }
    });
  }, observerOptions);

  document.querySelectorAll('.fade-in-up').forEach(el => {
    observer.observe(el);
  });

  // Add hover effect to cards
  document.querySelectorAll('.card-hover').forEach(card => {
    card.addEventListener('mouseenter', function(e) {
      const rect = this.getBoundingClientRect();
      const x = e.clientX - rect.left;
      const y = e.clientY - rect.top;

      this.style.setProperty('--x', x + 'px');
      this.style.setProperty('--y', y + 'px');
    });
  });

  // Add page load animation sequence
  window.addEventListener('load', function() {
    document.body.classList.add('loaded');
    
    // Trigger entrance animations with delays
    const animatedElements = document.querySelectorAll('[class*="animate-"]');
    animatedElements.forEach((el, index) => {
      if (!el.classList.contains('visible')) {
        setTimeout(() => {
          el.style.animationPlayState = 'running';
        }, index * 50);
      }
    });
  });

  // Set current year in footer
  const currentYearEl = document.getElementById('current-year');
  if (currentYearEl) currentYearEl.textContent = new Date().getFullYear();

  // Handle footer navigation links
  const footerLinks = document.querySelectorAll('.footer-nav-link');
  
  footerLinks.forEach(link => {
    link.addEventListener('click', function(e) {
      const section = this.getAttribute('data-section');
      
      if (section && window.location.pathname !== '/') {
        e.preventDefault();
        // Store the section to scroll to in sessionStorage
        sessionStorage.setItem('scrollToSection', section);
        window.location.href = '/';
      } else if (section && window.location.pathname === '/') {
        e.preventDefault();
        const element = document.getElementById(section);
        if (element) {
          const offset = 80; // Account for fixed nav
          const targetPosition = element.getBoundingClientRect().top + window.pageYOffset - offset;
          window.scrollTo({
            top: targetPosition,
            behavior: 'smooth'
          });
        }
      }
    });
  });

  // Check if we need to scroll to a section after navigation from another page
  const storedSection = sessionStorage.getItem('scrollToSection');
  
  if (storedSection) {
    sessionStorage.removeItem('scrollToSection');
    setTimeout(() => {
      const element = document.getElementById(storedSection);
      
      if (element) {
        const offset = 80; // Account for fixed nav
        const targetPosition = element.getBoundingClientRect().top + window.pageYOffset - offset;
        window.scrollTo({
          top: targetPosition,
          behavior: 'smooth'
        });
      }
    }, 100); // Small delay to ensure page is fully loaded
  }

});