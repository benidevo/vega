/**
 * Focus Management Utilities for Accessibility
 * Provides focus trap functionality for modals and keyboard navigation helpers
 */

(function() {
  'use strict';
  
  const deleteModalData = {
    itemToDelete: null
  };
  
  class FocusTrap {
  constructor(element, options = {}) {
    this.element = element;
    this.options = {
      initialFocus: options.initialFocus || null,
      returnFocus: options.returnFocus !== false,
      escapeDeactivates: options.escapeDeactivates !== false,
      ...options
    };
    
    this.focusableSelectors = [
      'a[href]:not([disabled])',
      'button:not([disabled])',
      'textarea:not([disabled])',
      'input:not([disabled])',
      'select:not([disabled])',
      '[tabindex]:not([tabindex="-1"])',
      '[contenteditable]:not([contenteditable="false"])'
    ];
    
    this.previouslyFocusedElement = null;
    this.handleKeyDown = this.handleKeyDown.bind(this);
    this.handleFocusIn = this.handleFocusIn.bind(this);
  }
  
  activate() {
    if (!this.element) return;
    this.previouslyFocusedElement = document.activeElement;
    this.updateFocusableElements();
    if (this.options.initialFocus) {
      const initialElement = this.element.querySelector(this.options.initialFocus);
      if (initialElement) {
        initialElement.focus();
      } else if (this.firstFocusableElement) {
        this.firstFocusableElement.focus();
      }
    } else if (this.firstFocusableElement) {
      this.firstFocusableElement.focus();
    }
    document.addEventListener('keydown', this.handleKeyDown);
    document.addEventListener('focusin', this.handleFocusIn);
    this.element.setAttribute('aria-modal', 'true');
    this.announceModal();
  }
  
  deactivate() {
    document.removeEventListener('keydown', this.handleKeyDown);
    document.removeEventListener('focusin', this.handleFocusIn);
    if (this.options.returnFocus && this.previouslyFocusedElement) {
      this.previouslyFocusedElement.focus();
    }
    this.element.removeAttribute('aria-modal');
  }
  
  updateFocusableElements() {
    const focusableElements = this.element.querySelectorAll(this.focusableSelectors.join(', '));
    this.focusableElements = Array.from(focusableElements).filter(el => {
      return !el.hasAttribute('disabled') && 
             !el.getAttribute('aria-hidden') && 
             el.offsetParent !== null;
    });
    
    this.firstFocusableElement = this.focusableElements[0];
    this.lastFocusableElement = this.focusableElements[this.focusableElements.length - 1];
  }
  
  handleKeyDown(e) {
    if (e.key === 'Tab') {
      this.handleTab(e);
    } else if (e.key === 'Escape' && this.options.escapeDeactivates) {
      this.options.onEscape && this.options.onEscape();
    }
  }
  
  handleTab(e) {
    this.updateFocusableElements();
    
    if (!this.focusableElements.length) {
      e.preventDefault();
      return;
    }
    
    if (e.shiftKey) {
      if (document.activeElement === this.firstFocusableElement) {
        e.preventDefault();
        this.lastFocusableElement.focus();
      }
    } else {
      if (document.activeElement === this.lastFocusableElement) {
        e.preventDefault();
        this.firstFocusableElement.focus();
      }
    }
  }
  
  handleFocusIn(e) {
    if (!this.element.contains(e.target)) {
      e.preventDefault();
      this.updateFocusableElements();
      if (this.firstFocusableElement) {
        this.firstFocusableElement.focus();
      }
    }
  }
  
  announceModal() {
    const announcer = document.getElementById('sr-announcements') || this.createAnnouncer();
    const modalTitle = this.element.getAttribute('aria-label') || 
                      this.element.querySelector('h2, h3, h4')?.textContent ||
                      'Dialog opened';
    
    announcer.textContent = modalTitle + '. Press Escape to close.';
  }
  
  createAnnouncer() {
    const existingAnnouncer = document.getElementById('sr-announcements');
    if (existingAnnouncer) {
      return existingAnnouncer;
    }
    
    const announcer = document.createElement('div');
    announcer.id = 'sr-announcements';
    announcer.className = 'sr-only';
    announcer.setAttribute('role', 'status');
    announcer.setAttribute('aria-live', 'polite');
    announcer.setAttribute('aria-atomic', 'true');
    document.body.appendChild(announcer);
    return announcer;
  }
}

  const FocusManager = {
  activeTrap: null,
  trapFocus(element, options = {}) {
    if (this.activeTrap) {
      this.activeTrap.deactivate();
    }
    
    this.activeTrap = new FocusTrap(element, options);
    this.activeTrap.activate();
    
    return this.activeTrap;
  },
  releaseFocus() {
    if (this.activeTrap) {
      this.activeTrap.deactivate();
      this.activeTrap = null;
    }
  },
  manageDropdown(trigger, menu, options = {}) {
    const isExpanded = trigger.getAttribute('aria-expanded') === 'true';
    
    if (!isExpanded) {
      trigger.setAttribute('aria-expanded', 'true');
      menu.hidden = false;
      const firstItem = menu.querySelector('a, button');
      if (firstItem) {
        firstItem.focus();
      }
      const items = Array.from(menu.querySelectorAll('a, button'));
      let currentIndex = 0;
      
      const handleKeyDown = (e) => {
        switch(e.key) {
          case 'ArrowDown':
            e.preventDefault();
            currentIndex = (currentIndex + 1) % items.length;
            items[currentIndex].focus();
            break;
            
          case 'ArrowUp':
            e.preventDefault();
            currentIndex = (currentIndex - 1 + items.length) % items.length;
            items[currentIndex].focus();
            break;
            
          case 'Home':
            e.preventDefault();
            currentIndex = 0;
            items[currentIndex].focus();
            break;
            
          case 'End':
            e.preventDefault();
            currentIndex = items.length - 1;
            items[currentIndex].focus();
            break;
            
          case 'Escape':
            e.preventDefault();
            this.closeDropdown(trigger, menu);
            trigger.focus();
            break;
            
          case 'Tab':
            this.closeDropdown(trigger, menu);
            break;
        }
      };
      
      menu.addEventListener('keydown', handleKeyDown);
      menu._keydownHandler = handleKeyDown;
      const handleClickOutside = (e) => {
        if (!menu.contains(e.target) && !trigger.contains(e.target)) {
          this.closeDropdown(trigger, menu);
        }
      };
      
      setTimeout(() => {
        document.addEventListener('click', handleClickOutside);
        menu._clickHandler = handleClickOutside;
      }, 0);
      
    } else {
      this.closeDropdown(trigger, menu);
    }
  },
  
  closeDropdown(trigger, menu) {
    trigger.setAttribute('aria-expanded', 'false');
    menu.hidden = true;
    if (menu._keydownHandler) {
      menu.removeEventListener('keydown', menu._keydownHandler);
      delete menu._keydownHandler;
    }
    
    if (menu._clickHandler) {
      document.removeEventListener('click', menu._clickHandler);
      delete menu._clickHandler;
    }
  },
  announce(message, priority = 'polite') {
    const announcer = document.getElementById('sr-announcements') || 
                     document.querySelector('[role="status"][aria-live]') ||
                     this.createAnnouncer();
    
    announcer.setAttribute('aria-live', priority);
    announcer.textContent = message;
    setTimeout(() => {
      announcer.textContent = '';
    }, 1000);
  },
  
  createAnnouncer() {
    const existingAnnouncer = document.getElementById('sr-announcements');
    if (existingAnnouncer) {
      return existingAnnouncer;
    }
    
    const announcer = document.createElement('div');
    announcer.id = 'sr-announcements';
    announcer.className = 'sr-only';
    announcer.setAttribute('role', 'status');
    announcer.setAttribute('aria-live', 'polite');
    announcer.setAttribute('aria-atomic', 'true');
    document.body.appendChild(announcer);
    return announcer;
  }
};

  window.FocusManager = {
    trapFocus: FocusManager.trapFocus.bind(FocusManager),
    releaseFocus: FocusManager.releaseFocus.bind(FocusManager),
    manageDropdown: FocusManager.manageDropdown.bind(FocusManager),
    closeDropdown: FocusManager.closeDropdown.bind(FocusManager),
    announce: FocusManager.announce.bind(FocusManager)
  };
  
  window.setDeleteItem = function(item) {
    if (item && typeof item === 'object') {
      deleteModalData.itemToDelete = Object.freeze({
        url: item.url || null,
        method: item.method || 'DELETE',
        data: item.data || null
      });
    } else {
      deleteModalData.itemToDelete = null;
    }
  };
  
  window.getDeleteItem = function() {
    return deleteModalData.itemToDelete;
  };
  
  window.clearDeleteItem = function() {
    deleteModalData.itemToDelete = null;
  };

  document.addEventListener('DOMContentLoaded', function() {
  const deleteModal = document.getElementById('delete-modal-shared');
  if (deleteModal) {
    const observer = new MutationObserver((mutations) => {
      mutations.forEach((mutation) => {
        if (mutation.type === 'attributes' && mutation.attributeName === 'class') {
          const isVisible = !deleteModal.classList.contains('hidden');
          
          if (isVisible) {
            window.FocusManager.trapFocus(deleteModal, {
              initialFocus: '#confirm-delete-btn',
              escapeDeactivates: true,
              onEscape: () => {
                deleteModal.classList.add('hidden');
              }
            });
          } else {
            window.FocusManager.releaseFocus();
          }
        }
      });
    });
    
    observer.observe(deleteModal, { attributes: true });
    
    window.addEventListener('beforeunload', () => {
      observer.disconnect();
    });
  }
  document.body.addEventListener('htmx:afterSwap', function(event) {
    if (event.detail.xhr.status >= 200 && event.detail.xhr.status < 300) {
      const target = event.detail.target;
      if (target.id === 'status-response' || 
          target.id === 'skills-response' || 
          target.id === 'notes-response' ||
          target.id === 'header-response') {
        window.FocusManager.announce('Changes saved successfully');
      }
    }
  });
  document.addEventListener('keydown', function(e) {
    if (['INPUT', 'TEXTAREA', 'SELECT'].includes(e.target.tagName)) {
      return;
    }
    switch(e.key) {
      case '?':
        if (e.shiftKey) {
          e.preventDefault();
          window.FocusManager.announce('Keyboard shortcuts: Escape to close dialogs');
        }
        break;
    }
  });
});

})();