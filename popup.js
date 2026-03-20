document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('scanForm');
    const resultsTable = document.getElementById('resultsTable');
    const resultsBody = document.getElementById('resultsBody');
    const errorMsg = document.getElementById('error-msg');
    const scanBtn = document.getElementById('scanBtn');

    // We are connecting to the local Go backend
    // Note: Since the backend uses self-signed certs (likely), the browser might block this request
    // unless the user has visited https://localhost:7373 first and accepted the risk.
    const BACKEND_URL = "https://localhost:7373/api/scan";

    form.addEventListener('submit', async (e) => {
        e.preventDefault();

        // Reset state
        errorMsg.style.display = 'none';
        resultsTable.style.display = 'none';
        resultsBody.innerHTML = '';
        scanBtn.disabled = true;
        scanBtn.value = "Scanning...";

        const target = document.getElementById('target').value.trim();
        const portsStr = document.getElementById('ports').value.trim();
        const timeout = document.getElementById('timeout').value;

        if (!target) {
            showError("Target is required");
            resetBtn();
            return;
        }

        try {
            // Build query string
            const params = new URLSearchParams({
                target: target,
                ports: portsStr,
                timeout: timeout
            });

            // Fetch from Go Backend
            const response = await fetch(`${BACKEND_URL}?${params.toString()}`);

            if (!response.ok) {
                throw new Error(`Server returned ${response.status}: ${await response.text()}`);
            }

            const results = await response.json();
            displayResults(results);

        } catch (err) {
            console.error(err);
            if (err.message.includes("Failed to fetch")) {
                showError("Failed to connect to backend. Is the Go server running? (Visit https://localhost:7373 to accept self-signed certs)");
            } else {
                showError(err.message);
            }
        } finally {
            resetBtn();
        }
    });

    function resetBtn() {
        scanBtn.disabled = false;
        scanBtn.value = "Scan";
    }

    function showError(msg) {
        errorMsg.textContent = "Error: " + msg;
        errorMsg.style.display = 'block';
    }

    function displayResults(results) {
        resultsBody.innerHTML = '';

        // Sort results by port
        results.sort((a, b) => a.port - b.port);

        if (results.length === 0) {
            const row = document.createElement('tr');
            row.innerHTML = `<td colspan="3">No ports found.</td>`;
            resultsBody.appendChild(row);
            return;
        }

        results.forEach((r, index) => {
            const row = document.createElement('tr');
            row.innerHTML = `
        <td>${index + 1}</td>
        <td>${r.port}</td>
        <td class="${r.status}">${r.status}</td>
      `;
            resultsBody.appendChild(row);
        });
        resultsTable.style.display = 'table';
    }
});
