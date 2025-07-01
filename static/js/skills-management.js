// Skills management functionality
function addSkill() {
    const input = document.getElementById('skills-input');
    const skill = input.value.trim();

    if (skill && !hasSkill(skill)) {
        createSkillTag(skill);
        input.value = '';
        updateSkills();
    }
}

function hasSkill(skill) {
    const tags = document.querySelectorAll('.skill-tag');
    return Array.from(tags).some(tag =>
        tag.textContent.trim().toLowerCase() === skill.toLowerCase()
    );
}

function createSkillTag(skill) {
    const container = document.getElementById('skills-tags');
    const tag = document.createElement('span');
    tag.className = 'skill-tag inline-flex items-center px-1 py-0.5 rounded text-xs font-medium bg-primary bg-opacity-20 text-primary border border-primary border-opacity-30';

    const skillText = document.createTextNode(skill);
    tag.appendChild(skillText);

    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'ml-1 inline-flex items-center justify-center w-4 h-4 text-primary hover:text-red-400 focus:outline-none';
    button.onclick = function() { removeSkillTag(this); };

    button.innerHTML = `
        <svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
        </svg>
    `;

    tag.appendChild(button);
    container.appendChild(tag);
}

function removeSkillTag(button) {
    const tag = button.closest('.skill-tag');
    if (tag) {
        tag.remove();
        updateSkills();
    }
}

function updateSkills() {
    const tags = document.querySelectorAll('.skill-tag');
    const skills = Array.from(tags).map(tag => tag.textContent.trim());
    document.getElementById('skills').value = skills.join(', ');
}