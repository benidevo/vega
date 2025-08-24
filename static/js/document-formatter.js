
window.DocumentFormatter = (function() {
  'use strict';

  function formatCVAsHTML(cvData) {
    if (typeof cvData === 'string') {
      try {
        cvData = JSON.parse(cvData);
      } catch (e) {
        console.error('Failed to parse CV data:', e);
        return cvData; // Return as-is if parsing fails
      }
    }

    let html = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Resume - ${cvData.personalInfo?.firstName || ''} ${cvData.personalInfo?.lastName || ''}</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Helvetica', Arial, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 40px 20px;
            background: white;
        }
        h1 { font-size: 28px; margin-bottom: 8px; color: #111; }
        h2 { 
            font-size: 16px; 
            margin-top: 24px; 
            margin-bottom: 12px; 
            padding-bottom: 4px;
            border-bottom: 2px solid #333;
            color: #111;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        h3 { font-size: 14px; margin-top: 16px; margin-bottom: 8px; color: #333; }
        h4 { font-size: 13px; margin-bottom: 4px; color: #333; }
        .header-section { margin-bottom: 24px; }
        .contact-info { 
            color: #666; 
            font-size: 12px;
            margin-bottom: 16px;
        }
        .title { 
            font-size: 14px;
            color: #444;
            margin-bottom: 8px;
        }
        .section { margin-bottom: 24px; }
        .experience-item, .education-item { 
            margin-bottom: 16px;
            page-break-inside: avoid;
        }
        .date { 
            color: #666; 
            font-size: 11px;
            margin-bottom: 4px;
        }
        .description { 
            font-size: 12px;
            line-height: 1.5;
            color: #444;
        }
        .skills { 
            font-size: 12px;
            line-height: 1.8;
        }
        ul {
            margin-left: 20px;
            margin-top: 4px;
        }
        li {
            font-size: 12px;
            margin-bottom: 4px;
        }
        @media print {
            body { padding: 0; }
            h2 { page-break-after: avoid; }
            .experience-item, .education-item { page-break-inside: avoid; }
        }
    </style>
</head>
<body>`;

    if (cvData.personalInfo) {
      const p = cvData.personalInfo;
      html += `
    <div class="header-section">
        <h1>${p.firstName || ''} ${p.lastName || ''}</h1>`;
      
      if (p.title) {
        html += `<div class="title">${p.title}</div>`;
      }

      const contactParts = [];
      if (p.location) contactParts.push(p.location);
      if (p.email) contactParts.push(p.email);
      if (p.phone) contactParts.push(p.phone);
      if (p.linkedin) contactParts.push(p.linkedin);
      
      if (contactParts.length > 0) {
        html += `<div class="contact-info">${contactParts.join(' | ')}</div>`;
      }

      if (p.summary) {
        html += `
        <div class="section">
            <h2>Professional Summary</h2>
            <p class="description">${p.summary}</p>
        </div>`;
      }
      html += `</div>`;
    }

    if (cvData.skills && cvData.skills.length > 0) {
      html += `
    <div class="section">
        <h2>Skills</h2>
        <div class="skills">${cvData.skills.join(' • ')}</div>
    </div>`;
    }

    if (cvData.workExperience && cvData.workExperience.length > 0) {
      html += `
    <div class="section">
        <h2>Work Experience</h2>`;
      
      cvData.workExperience.forEach(exp => {
        html += `
        <div class="experience-item">
            <h3>${exp.title} at ${exp.company}</h3>
            <div class="date">${exp.startDate} - ${exp.endDate}${exp.location ? ' | ' + exp.location : ''}</div>
            <div class="description">${formatDescription(exp.description)}</div>
        </div>`;
      });
      html += `</div>`;
    }

    if (cvData.education && cvData.education.length > 0) {
      html += `
    <div class="section">
        <h2>Education</h2>`;
      
      cvData.education.forEach(edu => {
        html += `
        <div class="education-item">
            <h3>${edu.degree}${edu.fieldOfStudy ? ' in ' + edu.fieldOfStudy : ''}</h3>
            <div>${edu.institution}</div>
            <div class="date">${edu.startDate} - ${edu.endDate}</div>
        </div>`;
      });
      html += `</div>`;
    }

    if (cvData.certifications && cvData.certifications.length > 0) {
      html += `
    <div class="section">
        <h2>Certifications</h2>`;
      
      cvData.certifications.forEach(cert => {
        html += `<div style="margin-bottom: 8px;">`;
        html += `<strong>${cert.name}</strong>`;
        if (cert.issuingOrg) html += ` - ${cert.issuingOrg}`;
        if (cert.issueDate) html += ` (${cert.issueDate})`;
        html += `</div>`;
      });
      html += `</div>`;
    }

    html += `
</body>
</html>`;

    return html;
  }

  function formatCoverLetterAsHTML(content, personalInfo) {
    if (!personalInfo) {
      personalInfo = {
        firstName: '',
        lastName: '',
        title: '',
        email: '',
        phone: '',
        location: ''
      };
    }

    const html = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Cover Letter</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'Helvetica', Arial, sans-serif;
            line-height: 1.8;
            color: #333;
            max-width: 800px;
            margin: 0 auto;
            padding: 40px 20px;
            background: white;
        }
        .header {
            margin-bottom: 30px;
            padding-bottom: 20px;
            border-bottom: 1px solid #e5e5e5;
        }
        h1 { 
            font-size: 24px; 
            margin-bottom: 8px; 
            color: #111; 
        }
        .title { 
            font-size: 14px;
            color: #666;
            margin-bottom: 8px;
        }
        .contact-info { 
            color: #666; 
            font-size: 12px;
        }
        .content {
            white-space: pre-wrap;
            font-size: 14px;
            line-height: 1.8;
        }
        p {
            margin-bottom: 16px;
        }
        @media print {
            body { padding: 20px; }
        }
    </style>
</head>
<body>`;

    if (personalInfo && (personalInfo.firstName || personalInfo.lastName)) {
      let headerHtml = `<div class="header">`;
      headerHtml += `<h1>${personalInfo.firstName || ''} ${personalInfo.lastName || ''}</h1>`;
      
      if (personalInfo.title) {
        headerHtml += `<div class="title">${personalInfo.title}</div>`;
      }

      const contactParts = [];
      if (personalInfo.location) contactParts.push(personalInfo.location);
      if (personalInfo.email) contactParts.push(personalInfo.email);
      if (personalInfo.phone) contactParts.push(personalInfo.phone);
      
      if (contactParts.length > 0) {
        headerHtml += `<div class="contact-info">${contactParts.join(' | ')}</div>`;
      }
      
      headerHtml += `</div>`;
      return html + headerHtml + `<div class="content">${escapeHtml(content)}</div></body></html>`;
    }

    return html + `<div class="content">${escapeHtml(content)}</div></body></html>`;
  }

  function formatDescription(text) {
    if (!text) return '';
    
    const lines = text.split('\n');
    let inList = false;
    let formatted = '';
    
    lines.forEach(line => {
      const trimmed = line.trim();
      if (trimmed.startsWith('•') || trimmed.startsWith('-') || trimmed.startsWith('*')) {
        if (!inList) {
          formatted += '<ul>';
          inList = true;
        }
        formatted += `<li>${trimmed.substring(1).trim()}</li>`;
      } else {
        if (inList) {
          formatted += '</ul>';
          inList = false;
        }
        if (trimmed) {
          formatted += `<p>${trimmed}</p>`;
        }
      }
    });
    
    if (inList) {
      formatted += '</ul>';
    }
    
    return formatted || text;
  }

  function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
  }

  return {
    formatCVAsHTML: formatCVAsHTML,
    formatCoverLetterAsHTML: formatCoverLetterAsHTML
  };
})();