/**
 * Shared JavaScript utilities for AI-generated content (CV and Cover Letter)
 * Used by templates/partials/cv_generator.html and cover_letter_generator.html
 */

window.AIHelpers = (function() {
    'use strict';

    // Common PDF generation utilities
    function loadHtml2PdfLibrary(callback, errorCallback) {
        if (typeof html2pdf !== 'undefined') {
            callback();
            return;
        }

        const script = document.createElement('script');
        script.src = 'https://cdnjs.cloudflare.com/ajax/libs/html2pdf.js/0.10.1/html2pdf.bundle.min.js';
        script.onload = callback;
        script.onerror = errorCallback || function() {
            alert('Failed to load PDF library. Please check your internet connection and try again.');
        };
        document.head.appendChild(script);
    }

    // Common PDF options
    function getPDFOptions(filename, format = 'a4') {
        return {
            margin: [0.5, 0.5, 0.5, 0.5],
            filename: filename,
            image: { type: 'jpeg', quality: 0.95 },
            html2canvas: {
                scale: 2,
                useCORS: true,
                letterRendering: true,
                allowTaint: false
            },
            jsPDF: {
                unit: 'in',
                format: format,
                orientation: 'portrait',
                compress: true
            }
        };
    }

    // Hide/show elements during PDF generation
    function hideElementsForPDF(selectors) {
        const elements = [];
        selectors.forEach(selector => {
            const element = document.querySelector(selector);
            if (element) {
                elements.push({ element, display: element.style.display });
                element.style.display = 'none';
            }
        });
        return elements;
    }

    function restoreElementsAfterPDF(hiddenElements) {
        hiddenElements.forEach(({ element, display }) => {
            element.style.display = display;
        });
    }

    // Common content validation
    function validateContentElement(elementId, contentType = 'content') {
        const element = document.getElementById(elementId);
        if (!element) {
            alert(`Error: ${contentType} not found. Please refresh the page and try again.`);
            return null;
        }
        return element;
    }

    // Generate PDF with error handling
    function generatePDFFromElement(elementId, filename, options = {}) {
        const element = validateContentElement(elementId);
        if (!element) return;

        const mergedOptions = { ...getPDFOptions(filename), ...options };

        return html2pdf()
            .set(mergedOptions)
            .from(element)
            .save()
            .catch(error => {
                console.error('PDF generation failed:', error);
                alert('PDF generation failed. Please try again.');
            });
    }

    // Copy to clipboard utility
    function copyToClipboard(text) {
        return navigator.clipboard.writeText(text).catch(err => {
            console.error('Failed to copy: ', err);
            alert('Failed to copy to clipboard');
        });
    }

    // Smooth scroll utility
    function scrollToElement(elementId, offset = 0) {
        const element = document.getElementById(elementId);
        if (element) {
            const elementPosition = element.offsetTop + offset;
            window.scrollTo({
                top: elementPosition,
                behavior: 'smooth'
            });
        }
    }

    // Mobile editing helpers
    function enableMobileEditing(containerId) {
        const container = document.getElementById(containerId);
        if (!container) return;

        // Add mobile-friendly classes and behaviors
        container.addEventListener('focus', function(e) {
            if (e.target.contentEditable === 'true') {
                // Scroll element into view on focus for mobile
                if (window.innerWidth < 768) {
                    setTimeout(() => {
                        e.target.scrollIntoView({
                            behavior: 'smooth',
                            block: 'center'
                        });
                    }, 300);
                }
            }
        }, true);
    }

    // Section management (for CV sections)
    function deleteSection(buttonElement) {
        if (!buttonElement) return;

        const section = buttonElement.closest('.resume-section');
        if (section && confirm('Are you sure you want to delete this section?')) {
            section.remove();
        }
    }

    // Modal utilities
    function showModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.remove('hidden');
            const firstInput = modal.querySelector('input, textarea, select');
            if (firstInput) {
                setTimeout(() => firstInput.focus(), 100);
            }
        }
    }

    function hideModal(modalId) {
        const modal = document.getElementById(modalId);
        if (modal) {
            modal.classList.add('hidden');
        }
    }

    // Validation utilities
    function validateRequiredFields(formElement) {
        const requiredFields = formElement.querySelectorAll('[required]');
        let isValid = true;

        requiredFields.forEach(field => {
            if (!field.value.trim()) {
                field.classList.add('border-red-500');
                isValid = false;
            } else {
                field.classList.remove('border-red-500');
            }
        });

        return isValid;
    }

    // Text utilities
    function sanitizeText(text) {
        return text.replace(/\s+/g, ' ').trim();
    }

    function truncateText(text, maxLength) {
        if (text.length <= maxLength) return text;
        return text.substring(0, maxLength) + '...';
    }

    return {
        // PDF utilities
        loadHtml2PdfLibrary,
        generatePDFFromElement,
        getPDFOptions,
        hideElementsForPDF,
        restoreElementsAfterPDF,

        // Content utilities
        validateContentElement,
        copyToClipboard,
        scrollToElement,
        enableMobileEditing,

        // Section management
        deleteSection,

        // Modal utilities
        showModal,
        hideModal,

        // Form utilities
        validateRequiredFields,

        // Text utilities
        sanitizeText,
        truncateText
    };
})();