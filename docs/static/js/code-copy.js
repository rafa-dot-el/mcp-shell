// Code copy button functionality for MCP Shell documentation
// Adds copy buttons to all code blocks and handles clipboard operations

(function() {
  'use strict';

  // Wait for DOM to be ready
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', initCodeCopy);
  } else {
    initCodeCopy();
  }

  function initCodeCopy() {
    // Find all code blocks (both .highlight and standalone pre elements)
    const codeBlocks = document.querySelectorAll('.highlight, pre:not(.highlight pre)');

    codeBlocks.forEach(function(block) {
      addCopyButton(block);
    });
  }

  function addCopyButton(codeBlock) {
    // Skip if button already exists
    if (codeBlock.querySelector('.copy-button')) {
      return;
    }

    // Create button wrapper
    const wrapper = document.createElement('div');
    wrapper.className = 'copy-button-wrapper';

    // Create button
    const button = document.createElement('button');
    button.className = 'copy-button';
    button.type = 'button';
    button.setAttribute('aria-label', 'Copy code to clipboard');

    // Button content with icon
    button.innerHTML = '<i class="fas fa-copy"></i><span class="copy-text">Copy</span>';

    // Add click handler
    button.addEventListener('click', function() {
      copyCode(codeBlock, button);
    });

    wrapper.appendChild(button);

    // Insert button at the beginning of the code block
    if (codeBlock.classList.contains('highlight')) {
      codeBlock.insertBefore(wrapper, codeBlock.firstChild);
    } else {
      // For standalone pre elements, wrap in a container
      const container = document.createElement('div');
      container.className = 'highlight';
      codeBlock.parentNode.insertBefore(container, codeBlock);
      container.appendChild(codeBlock);
      container.insertBefore(wrapper, container.firstChild);
    }
  }

  function copyCode(codeBlock, button) {
    // Find the code element
    let code = codeBlock.querySelector('code');
    if (!code) {
      code = codeBlock.querySelector('pre');
    }

    if (!code) {
      console.error('No code element found');
      return;
    }

    // Get the text content, removing line numbers if present
    let text = code.textContent || code.innerText;

    // Remove line numbers (they typically have specific patterns)
    text = text.replace(/^\s*\d+\s+/gm, '');

    // Copy to clipboard
    if (navigator.clipboard && navigator.clipboard.writeText) {
      // Modern clipboard API
      navigator.clipboard.writeText(text).then(function() {
        showCopySuccess(button);
      }).catch(function(err) {
        console.error('Failed to copy code: ', err);
        fallbackCopy(text, button);
      });
    } else {
      // Fallback for older browsers
      fallbackCopy(text, button);
    }
  }

  function fallbackCopy(text, button) {
    // Create temporary textarea
    const textarea = document.createElement('textarea');
    textarea.value = text;
    textarea.style.position = 'fixed';
    textarea.style.opacity = '0';
    textarea.style.pointerEvents = 'none';
    document.body.appendChild(textarea);

    try {
      textarea.select();
      const successful = document.execCommand('copy');
      if (successful) {
        showCopySuccess(button);
      } else {
        console.error('Fallback copy failed');
      }
    } catch (err) {
      console.error('Fallback copy error: ', err);
    } finally {
      document.body.removeChild(textarea);
    }
  }

  function showCopySuccess(button) {
    // Update button appearance
    const originalHTML = button.innerHTML;
    button.classList.add('copied');
    button.innerHTML = '<i class="fas fa-check"></i><span class="copy-text">Copied</span>';

    // Reset after 2 seconds
    setTimeout(function() {
      button.classList.remove('copied');
      button.innerHTML = originalHTML;
    }, 2000);
  }

  // Re-initialize when content changes (for dynamic content)
  if (typeof MutationObserver !== 'undefined') {
    const observer = new MutationObserver(function(mutations) {
      mutations.forEach(function(mutation) {
        if (mutation.addedNodes.length) {
          mutation.addedNodes.forEach(function(node) {
            if (node.nodeType === 1) { // Element node
              if (node.classList && (node.classList.contains('highlight') || node.tagName === 'PRE')) {
                addCopyButton(node);
              }
              // Check children
              const codeBlocks = node.querySelectorAll && node.querySelectorAll('.highlight, pre');
              if (codeBlocks) {
                codeBlocks.forEach(addCopyButton);
              }
            }
          });
        }
      });
    });

    observer.observe(document.body, {
      childList: true,
      subtree: true
    });
  }
})();
