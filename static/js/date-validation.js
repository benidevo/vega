// Date validation and UX improvements for work experience forms
document.addEventListener('DOMContentLoaded', function() {
    initializeDateValidation();
});

function initializeDateValidation() {
    const startDateInput = document.getElementById('start_date');
    const endDateInput = document.getElementById('end_date');
    const currentCheckbox = document.getElementById('current');

    if (!startDateInput) return;

    // Set max date to current month for both inputs
    const currentDate = new Date();
    const currentMonth = currentDate.getFullYear() + '-' + String(currentDate.getMonth() + 1).padStart(2, '0');
    startDateInput.max = currentMonth;
    if (endDateInput) endDateInput.max = currentMonth;

    // Add validation event listeners
    startDateInput.addEventListener('change', validateDates);
    if (endDateInput) endDateInput.addEventListener('change', validateDates);
    if (currentCheckbox) currentCheckbox.addEventListener('change', handleCurrentJobToggle);

    // Initial validation
    validateDates();
}

function validateDates() {
    const startDateInput = document.getElementById('start_date');
    const endDateInput = document.getElementById('end_date');
    const currentCheckbox = document.getElementById('current');

    if (!startDateInput || !endDateInput) return;

    const startDate = startDateInput.value;
    const endDate = endDateInput.value;

    clearDateErrors();

    if (currentCheckbox && currentCheckbox.checked) {
        return;
    }

    // Validate start date is not in the future
    if (startDate) {
        const currentDate = new Date();
        const currentMonth = currentDate.getFullYear() + '-' + String(currentDate.getMonth() + 1).padStart(2, '0');

        if (startDate > currentMonth) {
            showDateError(startDateInput, 'Start date cannot be in the future');
            return;
        }
    }

    // Validate end date is after start date
    if (startDate && endDate) {
        if (endDate < startDate) {
            showDateError(endDateInput, 'End date must be after start date');
            return;
        }

        // Calculate duration and show helpful info
        const duration = calculateDuration(startDate, endDate);
        if (duration) {
            showDateInfo(endDateInput, `Duration: ${duration}`);
        }
    }
}

function handleCurrentJobToggle() {
    const currentCheckbox = document.getElementById('current');
    const endDateInput = document.getElementById('end_date');

    if (!currentCheckbox || !endDateInput) return;

    if (currentCheckbox.checked) {
        endDateInput.disabled = true;
        endDateInput.value = '';
        clearDateErrors();

        // Show duration from start to current
        const startDate = document.getElementById('start_date').value;
        if (startDate) {
            const currentDate = new Date();
            const currentMonth = currentDate.getFullYear() + '-' + String(currentDate.getMonth() + 1).padStart(2, '0');
            const duration = calculateDuration(startDate, currentMonth);
            if (duration) {
                showDateInfo(endDateInput, `Current duration: ${duration}`);
            }
        }
    } else {
        endDateInput.disabled = false;
        clearDateErrors();
    }
}

function calculateDuration(startDate, endDate) {
    if (!startDate || !endDate) return null;

    const [startYear, startMonth] = startDate.split('-').map(Number);
    const [endYear, endMonth] = endDate.split('-').map(Number);

    const startMonthTotal = startYear * 12 + startMonth;
    const endMonthTotal = endYear * 12 + endMonth;

    const totalMonths = endMonthTotal - startMonthTotal;

    if (totalMonths < 0) return null;

    const years = Math.floor(totalMonths / 12);
    const months = totalMonths % 12;

    if (years === 0 && months === 0) {
        return 'Less than 1 month';
    } else if (years === 0) {
        return `${months} month${months !== 1 ? 's' : ''}`;
    } else if (months === 0) {
        return `${years} year${years !== 1 ? 's' : ''}`;
    } else {
        return `${years} year${years !== 1 ? 's' : ''} ${months} month${months !== 1 ? 's' : ''}`;
    }
}

function showDateError(input, message) {
    clearDateErrors();

    const errorDiv = document.createElement('div');
    errorDiv.className = 'date-error mt-1 text-sm text-red-400';
    errorDiv.textContent = message;

    input.parentElement.appendChild(errorDiv);
    input.classList.add('border-red-500');
}

function showDateInfo(input, message) {
    clearDateInfo();

    const infoDiv = document.createElement('div');
    infoDiv.className = 'date-info mt-1 text-sm text-primary';
    infoDiv.textContent = message;

    input.parentElement.appendChild(infoDiv);
}

function clearDateErrors() {
    const errors = document.querySelectorAll('.date-error');
    errors.forEach(error => error.remove());

    const inputs = document.querySelectorAll('input[type="month"]');
    inputs.forEach(input => input.classList.remove('border-red-500'));
}

function clearDateInfo() {
    const infos = document.querySelectorAll('.date-info');
    infos.forEach(info => info.remove());
}

// Add mobile-friendly date input handling
function enhanceMobileDateExperience() {
    const dateInputs = document.querySelectorAll('input[type="month"]');

    dateInputs.forEach(input => {
        input.style.fontSize = '16px'; // Prevents zoom on iOS

        input.addEventListener('touchstart', function() {
            this.focus();
        });

        // Improve mobile keyboard experience
        input.addEventListener('focus', function() {
            // Scroll input into view on mobile
            setTimeout(() => {
                this.scrollIntoView({
                    behavior: 'smooth',
                    block: 'center'
                });
            }, 300);
        });
    });
}

// Initialize mobile enhancements
document.addEventListener('DOMContentLoaded', enhanceMobileDateExperience);
