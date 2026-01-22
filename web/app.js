// OpenConvert Frontend Application

let selectedFiles = [];

// Initialize drag and drop
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
        selectedFiles = [...selectedFiles, ...files];
        updateFileList();
        convertButton.disabled = selectedFiles.length === 0;
    }

    function updateFileList() {
        if (selectedFiles.length === 0) {
            fileList.style.display = 'none';
            return;
        }

        fileList.style.display = 'block';
        fileItems.innerHTML = selectedFiles.map((file, index) => `
            <li class="file-item">
                <span>${file.name}</span>
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

