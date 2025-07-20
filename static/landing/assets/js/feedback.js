document.addEventListener('DOMContentLoaded', function() {
  const slider = document.getElementById('setup-difficulty');
  if (slider) {
    const initialPercent = ((slider.value - slider.min) / (slider.max - slider.min)) * 100;
    slider.style.setProperty('--value', initialPercent + '%');
  }
});

function toggleFeedbackForm() {
  const container = document.getElementById('feedback-form-container');
  const icon = document.getElementById('expand-icon');
  const slider = document.getElementById('setup-difficulty');
  
  container.classList.toggle('hidden');
  icon.classList.toggle('rotate-180');
  
  if (!container.classList.contains('hidden') && slider) {
    slider.style.setProperty('--value', ((slider.value - 1) / 9 * 100) + '%');
  }
}

function updateSliderValue(slider) {
  document.getElementById('difficulty-value').innerText = slider.value;
  slider.style.setProperty('--value', ((slider.value - 1) / 9 * 100) + '%');
}

async function submitFeedback(event) {
  event.preventDefault();
  
  const form = event.target;
  const formData = new FormData(form);
  const messageDiv = document.getElementById('form-message');
  
  try {
    messageDiv.innerHTML = '<div class="bg-blue-500/10 border border-blue-500/20 rounded-lg p-4 text-blue-400">Sending feedback...</div>';
    
    const urlParams = new URLSearchParams(formData);
    
    const response = await fetch('https://us-central1-vega-ai-live.cloudfunctions.net/vega-landing-api?action=feedback', {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: urlParams
    });
    
    if (!response.ok) {
      console.warn('Feedback submission failed:', response.status, response.statusText);
    }
    
  } catch (error) {
    console.warn('Feedback submission error:', error);
  }
  
  messageDiv.innerHTML = '<div class="bg-green-500/10 border border-green-500/20 rounded-lg p-4 text-green-400"><div class="flex items-center gap-2"><svg class="w-5 h-5" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path></svg><span>Thank you for your feedback!</span></div></div>';
  
  form.reset();
  document.getElementById('difficulty-value').innerText = '5';
  setTimeout(() => {
    document.getElementById('feedback-form-container').classList.add('hidden');
    document.getElementById('expand-icon').classList.remove('rotate-180');
  }, 3000);
}