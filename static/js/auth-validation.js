// Client-side validation for login form
function validateLoginForm(event) {
  const username = document.getElementById('username').value;
  const password = document.getElementById('password').value;
  const errorContainer = document.getElementById('client-error');
  const errorMessage = document.getElementById('client-error-message');
  const formResponse = document.getElementById('form-response');

  let message = '';
  let focusElement = null;

  if (!username || username.length < 3 || username.length > 50) {
    if (!username) {
      message = 'Username is required';
    } else if (username.length < 3) {
      message = 'Username must be at least 3 characters long';
    } else {
      message = 'Username must be no more than 50 characters long';
    }
    focusElement = document.getElementById('username');
  }

  else if (!password || password.length < 8 || password.length > 64) {
    if (!password) {
      message = 'Password is required';
    } else if (password.length < 8) {
      message = 'Password must be at least 8 characters long';
    } else {
      message = 'Password must be no more than 64 characters long';
    }
    focusElement = document.getElementById('password');
  }

  // If validation failed, show error and prevent form submission
  if (message) {
    errorMessage.textContent = message;
    errorContainer.classList.remove('hidden');
    if (formResponse) {
      formResponse.classList.add('hidden');
    }
    if (focusElement) {
      focusElement.focus();
    }
    event.preventDefault();
    return false;
  }

  // Hide error if validation passed
  errorContainer.classList.add('hidden');
  return true;
}

// Hide client errors after successful HTMX request
document.addEventListener('htmx:afterRequest', function() {
  const errorContainer = document.getElementById('client-error');
  if (errorContainer) {
    errorContainer.classList.add('hidden');
  }
});
