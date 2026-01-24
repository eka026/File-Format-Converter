// File Format Converter Frontend Application

let selectedFiles = [];

// Supported file types for FR-01
const SUPPORTED_EXTENSIONS = ['.xlsx'];
const XLSX_MIME_TYPES = [
    'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
];

// Validates if a file is a supported .xlsx file
function isValidXlsxFile(file) {
    const fileName = file.name.toLowerCase();
    const hasValidExtension = SUPPORTED_EXTENSIONS.some(ext => fileName.endsWith(ext));
    const hasValidMimeType = XLSX_MIME_TYPES.includes(file.type) || file.type === '';
    return hasValidExtension && hasValidMimeType;
}

// Gets file extension from filename
function getFileExtension(filename) {
    return filename.slice(filename.lastIndexOf('.')).toLowerCase();
}

document.addEventListener('DOMContentLoaded', () => {
    const dropZone = document.getElementById('dropZone');
    const fileInput = document.getElementById('fileInput');
    const convertButton = document.getElementById('convertButton');
    const fileList = document.getElementById('fileList');
    const fileItems = document.getElementById('fileItems');

    // Drag and drop handlers
    dropZone.addEventListener('click', () => fileInput.click());
    dropZone.addEventListener('dragover', handleDragOver);
    dropZone.addEventListener('dragleave', handleDragLeave);
    dropZone.addEventListener('drop', handleDrop);
    fileInput.addEventListener('change', handleFileSelect);

    // Convert button handler
    convertButton.addEventListener('click', handleConvert);

    function handleDragOver(e) {
        e.preventDefault();
        dropZone.classList.add('dragover');
    }

    function handleDragLeave(e) {
        e.preventDefault();
        dropZone.classList.remove('dragover');
    }

    function handleDrop(e) {
        e.preventDefault();
        dropZone.classList.remove('dragover');
        const files = Array.from(e.dataTransfer.files);
        addFiles(files);
    }

    function handleFileSelect(e) {
        const files = Array.from(e.target.files);
        addFiles(files);
    }

    function addFiles(files) {
        const validFiles = [];
        const invalidFiles = [];

        files.forEach(file => {
            if (isValidXlsxFile(file)) {
                validFiles.push(file);
            } else {
                invalidFiles.push(file);
            }
        });

        // Add valid files to selection
        selectedFiles = [...selectedFiles, ...validFiles];
        updateFileList();
        convertButton.disabled = selectedFiles.length === 0;

        // Show error for invalid files
        if (invalidFiles.length > 0) {
            showValidationError(invalidFiles);
        }
    }

    function showValidationError(invalidFiles) {
        const results = document.getElementById('results');
        const fileNames = invalidFiles.map(f => f.name).join(', ');
        results.innerHTML = `
            <div class="result-item error">
                <strong>Invalid file type</strong>
                <p>The following files are not supported: ${fileNames}</p>
                <p>Supported format: .xlsx (Excel files)</p>
            </div>
        `;
    }

    function updateFileList() {
        if (selectedFiles.length === 0) {
            fileList.style.display = 'none';
            return;
        }

        fileList.style.display = 'block';
        fileItems.innerHTML = selectedFiles.map((file, index) => `
            <li class="file-item">
                <span class="file-icon xlsx-icon">📊</span>
                <span class="file-name">${file.name}</span>
                <span class="file-type-badge">XLSX</span>
                <button onclick="removeFile(${index})">Remove</button>
            </li>
        `).join('');
    }

    window.removeFile = (index) => {
        selectedFiles.splice(index, 1);
        updateFileList();
        convertButton.disabled = selectedFiles.length === 0;
    };

    async function handleConvert() {
        const targetFormat = document.getElementById('targetFormat').value;
        const progressContainer = document.getElementById('progressContainer');
        const progressFill = document.getElementById('progressFill');
        const progressText = document.getElementById('progressText');
        const results = document.getElementById('results');

        // Show and initialize progress bar
        console.log('=== Starting Conversion Process ===');
        console.log('Selected files:', selectedFiles.length);
        progressContainer.style.display = 'block';
        progressFill.style.width = '0%';
        progressText.textContent = '0%';
        results.innerHTML = '';
        convertButton.disabled = true;
        
        // Force a repaint to ensure progress bar is visible
        await new Promise(resolve => setTimeout(resolve, 10));
        console.log('Progress bar initialized at 0%');

        // Check if Wails API is available
        if (typeof window.go === 'undefined' || !window.go.gui || !window.go.gui.App) {
            // Fallback: show error if Wails is not available
            results.innerHTML = `
                <div class="result-item error">
                    <strong>Error</strong>
                    <p>Backend connection not available. Please ensure the application is running properly.</p>
                    <p style="font-size: 0.85em; margin-top: 8px;">
                        window.go: ${typeof window.go}, window.go.gui: ${typeof window.go?.gui}
                    </p>
                </div>
            `;
            convertButton.disabled = false;
            progressContainer.style.display = 'none';
            return;
        }

        const app = window.go.gui.App;
        const conversionResults = [];
        
        // Verify required methods exist
        if (!app.ConvertFile || !app.SaveFileFromBytes) {
            results.innerHTML = `
                <div class="result-item error">
                    <strong>Error</strong>
                    <p>Required backend methods are not available. Please rebuild the application.</p>
                    <p style="font-size: 0.85em; margin-top: 8px;">
                        Missing: ${!app.ConvertFile ? 'ConvertFile ' : ''}${!app.SaveFileFromBytes ? 'SaveFileFromBytes' : ''}
                    </p>
                </div>
            `;
            convertButton.disabled = false;
            progressContainer.style.display = 'none';
            return;
        }

        try {
            // Check if there are files to convert
            if (selectedFiles.length === 0) {
                results.innerHTML = `
                    <div class="result-item error">
                        <strong>No Files Selected</strong>
                        <p>Please select at least one file to convert.</p>
                    </div>
                `;
                convertButton.disabled = false;
                progressContainer.style.display = 'none';
                return;
            }

            // Update progress - starting (already initialized to 0% above)
            progressFill.style.width = '2%';
            progressText.textContent = '2%';
            await new Promise(resolve => setTimeout(resolve, 100)); // Ensure UI updates

            // Update progress - initializing
            progressFill.style.width = '5%';
            progressText.textContent = '5%';
            await new Promise(resolve => setTimeout(resolve, 100));

            // Convert each file
            for (let i = 0; i < selectedFiles.length; i++) {
                const file = selectedFiles[i];
                
                // Update progress - preparing file
                const prepProgress = Math.floor((i / selectedFiles.length) * 30) + 5;
                progressFill.style.width = prepProgress + '%';
                progressText.textContent = prepProgress + '%';
                await new Promise(resolve => setTimeout(resolve, 50));

                let filePath = file.path || file.name;
                let isTempInputFile = false;
                
                // If we have a browser File object (from drag-and-drop), save it to temp first
                if (file instanceof File && !file.path) {
                    try {
                        // Update progress - reading file
                        progressFill.style.width = (prepProgress + 5) + '%';
                        progressText.textContent = (prepProgress + 5) + '%';
                        
                        // Read file as array buffer
                        const arrayBuffer = await file.arrayBuffer();
                        const fileData = new Uint8Array(arrayBuffer);
                        
                        // Update progress - saving file
                        progressFill.style.width = (prepProgress + 10) + '%';
                        progressText.textContent = (prepProgress + 10) + '%';
                        
                        // Save to temp location via backend
                        filePath = await app.SaveFileFromBytes(file.name, Array.from(fileData));
                        isTempInputFile = true; // Mark as temp file for cleanup
                        console.log('File saved to temp:', filePath);
                    } catch (error) {
                        console.error('Error saving file:', error);
                        conversionResults.push({
                            success: false,
                            error: `Failed to save file: ${error.message}`
                        });
                        continue;
                    }
                }
                
                // Update progress - starting conversion
                const conversionStartProgress = Math.floor((i / selectedFiles.length) * 60) + 35;
                progressFill.style.width = conversionStartProgress + '%';
                progressText.textContent = conversionStartProgress + '%';
                await new Promise(resolve => setTimeout(resolve, 50));
                
                // Perform conversion (save dialog disabled due to WebSocket issues)
                // PDF will be saved to default location (same directory as input, or Downloads folder)
                console.log('Starting conversion:', filePath, 'to', targetFormat);
                console.log('Progress before conversion:', conversionStartProgress + '%');
                try {
                    // Convert with empty path - backend will use default location
                    const result = await app.ConvertFile(filePath, targetFormat);
                    console.log('Conversion result:', JSON.stringify(result, null, 2));
                    if (!result.success) {
                        console.error('Conversion failed:', result.error);
                    }
                    conversionResults.push(result);
                    
                    // Clean up temp input file after successful conversion
                    if (isTempInputFile) {
                        try {
                            await app.CleanupTempInputFile(filePath);
                            console.log('Cleaned up temp input file:', filePath);
                        } catch (cleanupError) {
                            console.warn('Failed to cleanup temp input file:', cleanupError);
                        }
                    }
                } catch (error) {
                    console.error('Conversion exception caught:', error);
                    console.error('Error details:', {
                        message: error.message,
                        stack: error.stack,
                        name: error.name
                    });
                    conversionResults.push({
                        success: false,
                        error: error.message || 'Unknown conversion error'
                    });
                    
                    // Clean up temp input file even on error
                    if (isTempInputFile) {
                        try {
                            await app.CleanupTempInputFile(filePath);
                            console.log('Cleaned up temp input file after error:', filePath);
                        } catch (cleanupError) {
                            console.warn('Failed to cleanup temp input file:', cleanupError);
                        }
                    }
                }
                
                // Update progress after conversion
                const postConversionProgress = Math.floor(((i + 1) / selectedFiles.length) * 60) + 35;
                progressFill.style.width = postConversionProgress + '%';
                progressText.textContent = postConversionProgress + '%';
                await new Promise(resolve => setTimeout(resolve, 50)); // Ensure UI updates
            }

            // Complete progress
            progressFill.style.width = '95%';
            progressText.textContent = '95%';
            await new Promise(resolve => setTimeout(resolve, 100));
            
            progressFill.style.width = '100%';
            progressText.textContent = '100%';
            await new Promise(resolve => setTimeout(resolve, 100)); // Ensure UI updates

            // Show results with download/open options
            const successCount = conversionResults.filter(r => r.success).length;
            const failedCount = conversionResults.length - successCount;

            let resultHTML = `
                <div class="result-item ${successCount > 0 ? 'success' : 'error'}">
                    <strong>Conversion ${successCount > 0 ? 'Complete' : 'Failed'}!</strong>
                    <p>Successfully converted ${successCount} file(s) to ${targetFormat.toUpperCase()}</p>
                    ${failedCount > 0 ? `<p>Failed to convert ${failedCount} file(s)</p>` : ''}
                </div>
            `;

            // Add download/open buttons for successful conversions
            conversionResults.forEach((result, index) => {
                if (result.success && result.outputPath) {
                    // Escape the path properly for use in HTML/JavaScript
                    // Replace backslashes with forward slashes for display
                    const displayPath = result.outputPath.replace(/\\/g, '/');
                    const fileName = result.outputPath.split(/[/\\]/).pop();
                    // Escape HTML special characters for safe display
                    const safeDisplayPath = displayPath.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
                    const safeFileName = fileName.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
                    // Use data attribute to store the full path safely
                    const safePathAttr = result.outputPath.replace(/\\/g, '\\\\').replace(/'/g, "\\'").replace(/"/g, '&quot;');
                    resultHTML += `
                        <div class="result-item success file-result" data-file-path="${safePathAttr}">
                            <p><strong>File ${index + 1}:</strong> ${safeFileName}</p>
                            <p class="file-path">${safeDisplayPath}</p>
                            <div class="file-actions">
                                <button class="action-button open-pdf-btn">Open PDF</button>
                                <button class="action-button show-folder-btn">Show in Folder</button>
                            </div>
                        </div>
                    `;
                } else if (!result.success) {
                    resultHTML += `
                        <div class="result-item error">
                            <p><strong>File ${index + 1} failed:</strong> ${(result.error || 'Unknown error').replace(/</g, '&lt;').replace(/>/g, '&gt;')}</p>
                        </div>
                    `;
                }
            });

            results.innerHTML = resultHTML;

            // Attach event listeners to the buttons using event delegation
            results.querySelectorAll('.open-pdf-btn').forEach(button => {
                button.addEventListener('click', function() {
                    const fileResult = this.closest('.file-result');
                    const filePath = fileResult.getAttribute('data-file-path');
                    if (filePath) {
                        openFile(filePath);
                    }
                });
            });

            results.querySelectorAll('.show-folder-btn').forEach(button => {
                button.addEventListener('click', function() {
                    const fileResult = this.closest('.file-result');
                    const filePath = fileResult.getAttribute('data-file-path');
                    if (filePath) {
                        showFileLocation(filePath);
                    }
                });
            });

        } catch (error) {
            console.error('Conversion error:', error);
            // Update progress to show something happened, even on error
            progressFill.style.width = '50%';
            progressText.textContent = '50%';
            await new Promise(resolve => setTimeout(resolve, 100));
            
            // Complete progress even on error
            progressFill.style.width = '100%';
            progressText.textContent = '100%';
            await new Promise(resolve => setTimeout(resolve, 100));
            
            results.innerHTML = `
                <div class="result-item error">
                    <strong>Conversion Error</strong>
                    <p>${error.message || 'An unexpected error occurred during conversion'}</p>
                    <p style="font-size: 0.85em; margin-top: 8px; color: var(--text-tertiary);">
                        Check the browser console for more details.
                    </p>
                </div>
            `;
        } finally {
            convertButton.disabled = false;
            selectedFiles = [];
            updateFileList();
        }
    }

    // Helper function to open a file
    window.openFile = async function(filePath) {
        try {
            console.log('Opening file:', filePath);
            
            // Verify the path is valid
            if (!filePath || filePath.trim() === '') {
                alert('Invalid file path');
                return;
            }
            
            const app = window.go.gui.App;
            await app.OpenFile(filePath);
        } catch (error) {
            console.error('Failed to open file:', error);
            alert('Failed to open file: ' + (error.message || error));
        }
    };

    // Helper function to show file location
    window.showFileLocation = function(filePath) {
        try {
            // Extract directory path
            const pathParts = filePath.split(/[/\\]/);
            const fileName = pathParts[pathParts.length - 1];
            const directory = pathParts.slice(0, -1).join('\\');
            
            alert(`File saved to:\n${directory}\n\nFilename: ${fileName}`);
        } catch (error) {
            console.error('Failed to show file location:', error);
            alert('Failed to show file location: ' + (error.message || error));
        }
    };
});
