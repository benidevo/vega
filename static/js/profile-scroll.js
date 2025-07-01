// Profile page scroll to section functionality
document.addEventListener('DOMContentLoaded', function() {
    const urlParams = new URLSearchParams(window.location.search);
    const scrollSection = urlParams.get('scroll');

    if (scrollSection) {
        const sectionId = scrollSection + '-section';
        const element = document.getElementById(sectionId);

        if (element) {
            // Small delay to ensure page is fully loaded
            setTimeout(() => {
                element.scrollIntoView({
                    behavior: 'smooth',
                    block: 'start'
                });

                // Clean up URL without anchor
                const cleanUrl = window.location.pathname;
                window.history.replaceState({}, document.title, cleanUrl);
            }, 100);
        }
    }
});
