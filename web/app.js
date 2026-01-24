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

        progressContainer.style.display = 'block';
        results.innerHTML = '';
        convertButton.disabled = true;

        // Simulate conversion progress (will be replaced with actual API calls)
        for (let i = 0; i <= 100; i += 10) {
            progressFill.style.width = i + '%';
            progressText.textContent = i + '%';
            await new Promise(resolve => setTimeout(resolve, 200));
        }

        // Show results
        results.innerHTML = `
            <div class="result-item success">
                <strong>Conversion Complete!</strong>
                <p>Converted ${selectedFiles.length} file(s) to ${targetFormat.toUpperCase()}</p>
            </div>
        `;

        convertButton.disabled = false;
        selectedFiles = [];
        updateFileList();
    }
});
