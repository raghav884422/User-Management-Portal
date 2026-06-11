// APP STATE
const state = {
    currentPage: 1,
    pageSize: 10,
    totalPages: 1,
    totalUsers: 0,
    isEditing: false,
    editingUserId: null,
    deleteUserCandidateId: null
};

// DOM ELEMENTS
const dom = {
    form: document.getElementById('user-form'),
    userIdInput: document.getElementById('user-id'),
    nameInput: document.getElementById('user-name'),
    dobInput: document.getElementById('user-dob'),
    formTitle: document.getElementById('form-title'),
    formIcon: document.getElementById('form-icon'),
    btnSubmitText: document.getElementById('btn-text'),
    btnCancel: document.getElementById('btn-cancel'),
    tableBody: document.getElementById('users-table-body'),
    totalUsersBadge: document.getElementById('total-users-badge'),
    pageSizeSelect: document.getElementById('page-size'),
    btnPrev: document.getElementById('btn-prev'),
    btnNext: document.getElementById('btn-next'),
    paginationInfo: document.getElementById('pagination-info'),
    apiStatusWidget: document.getElementById('api-status-widget'),
    toastContainer: document.getElementById('toast-container'),
    nameError: document.getElementById('name-error'),
    dobError: document.getElementById('dob-error'),
    deleteModalBackdrop: document.getElementById('delete-modal-backdrop'),
    deleteModalClose: document.getElementById('delete-modal-close'),
    deleteModalCancel: document.getElementById('delete-modal-cancel'),
    deleteModalConfirm: document.getElementById('delete-modal-confirm'),
    deleteModalUser: document.getElementById('delete-modal-user')
};

// INITIALIZATION
document.addEventListener('DOMContentLoaded', () => {
    loadUsers();
    startHealthCheck();
    setupEventListeners();
});

// LISTENERS
function setupEventListeners() {
    // Form submit
    dom.form.addEventListener('submit', handleFormSubmit);

    // Cancel edit
    dom.btnCancel.addEventListener('click', resetForm);

    // Page size change
    dom.pageSizeSelect.addEventListener('change', (e) => {
        state.pageSize = parseInt(e.target.value, 10);
        state.currentPage = 1;
        loadUsers();
    });

    // Pagination buttons
    dom.btnPrev.addEventListener('click', () => {
        if (state.currentPage > 1) {
            state.currentPage--;
            loadUsers();
        }
    });

    dom.btnNext.addEventListener('click', () => {
        if (state.currentPage < state.totalPages) {
            state.currentPage++;
            loadUsers();
        }
    });

    dom.deleteModalCancel.addEventListener('click', closeDeleteModal);
    dom.deleteModalClose.addEventListener('click', closeDeleteModal);
    dom.deleteModalConfirm.addEventListener('click', confirmDeleteUser);
    dom.deleteModalBackdrop.addEventListener('click', (event) => {
        if (event.target === dom.deleteModalBackdrop) {
            closeDeleteModal();
        }
    });
}

// TOAST NOTIFICATIONS
function showToast(message, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast toast-${type}`;
    
    let iconClass = 'fa-info-circle';
    if (type === 'success') iconClass = 'fa-check-circle';
    if (type === 'error') iconClass = 'fa-exclamation-circle';
    
    toast.innerHTML = `
        <i class="fa-solid ${iconClass}"></i>
        <span class="toast-message">${message}</span>
    `;
    
    dom.toastContainer.appendChild(toast);
    
    // Auto remove
    setTimeout(() => {
        toast.style.opacity = '0';
        toast.style.transform = 'translateY(20px)';
        setTimeout(() => toast.remove(), 300);
    }, 4000);
}

// HEALTH CHECK POLLING
function startHealthCheck() {
    const check = async () => {
        try {
            const res = await fetch('/health');
            const data = await res.json();
            
            const dot = dom.apiStatusWidget.querySelector('.status-dot');
            const text = dom.apiStatusWidget.querySelector('.status-text');
            
            if (res.ok && data.status === 'ok') {
                dot.className = 'status-dot online pulsing';
                text.textContent = 'API Server Online';
            } else {
                throw new Error('Unhealthy status');
            }
        } catch (err) {
            const dot = dom.apiStatusWidget.querySelector('.status-dot');
            const text = dom.apiStatusWidget.querySelector('.status-text');
            dot.className = 'status-dot offline pulsing';
            text.textContent = 'API Connection Offline';
        }
    };

    check();
    setInterval(check, 10000); // Check every 10 seconds
}

// LOAD USERS FROM API
async function loadUsers() {
    showTableLoading();
    try {
        const url = `/users?page=${state.currentPage}&page_size=${state.pageSize}`;
        const res = await fetch(url);
        if (!res.ok) throw new Error('Failed to retrieve registry catalog');
        
        const responseData = await res.json();
        const users = responseData.data || [];
        
        state.totalUsers = responseData.total || 0;
        state.totalPages = responseData.total_pages || 1;
        state.currentPage = responseData.page || 1;
        
        updatePaginationUI();
        renderUsersTable(users);
    } catch (err) {
        showToast(err.message, 'error');
        showTableError(err.message);
    }
}

// RENDER TABLE ROWS
function renderUsersTable(users) {
    if (users.length === 0) {
        dom.tableBody.innerHTML = `
            <tr>
                <td colspan="5" class="table-empty">
                    <i class="fa-solid fa-folder-open" style="font-size: 2.5rem; color: var(--text-muted); margin-bottom: 1rem; display: block;"></i>
                    <span>No users currently registered in the database.</span>
                </td>
            </tr>
        `;
        return;
    }

    dom.tableBody.innerHTML = '';
    users.forEach(user => {
        const initials = user.name.split(' ').map(n => n[0]).join('').substring(0, 2).toUpperCase();
        const formattedDob = formatDateString(user.dob);
        
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>#${user.id}</td>
            <td>
                <div class="profile-cell">
                    <div class="avatar">${initials}</div>
                    <strong>${escapeHTML(user.name)}</strong>
                </div>
            </td>
            <td>${formattedDob}</td>
            <td><span class="age-badge">${user.age} yrs</span></td>
            <td>
                <div class="action-buttons">
                    <button class="btn-icon btn-icon-edit" onclick="initiateEditUser(${user.id}, '${escapeQuote(user.name)}', '${user.dob.split('T')[0]}')" title="Edit Profile">
                        <i class="fa-solid fa-pen"></i>
                    </button>
                    <button class="btn-icon btn-icon-delete" onclick="handleDeleteUser(${user.id}, '${escapeQuote(user.name)}')" title="Delete User">
                        <i class="fa-solid fa-trash-can"></i>
                    </button>
                </div>
            </td>
        `;
        dom.tableBody.appendChild(row);
    });
}

// FORM SUBMISSION HANDLER
async function handleFormSubmit(e) {
    e.preventDefault();
    
    // Reset errors
    dom.nameError.style.display = 'none';
    dom.dobError.style.display = 'none';
    
    const name = dom.nameInput.value.trim();
    const dob = dom.dobInput.value;
    
    let hasError = false;
    
    if (!name) {
        dom.nameError.textContent = 'Full Name is required';
        dom.nameError.style.display = 'block';
        hasError = true;
    }
    
    if (!dob) {
        dom.dobError.textContent = 'Date of Birth is required';
        dom.dobError.style.display = 'block';
        hasError = true;
    } else {
        const dobDate = new Date(dob);
        const today = new Date();
        if (dobDate > today) {
            dom.dobError.textContent = 'Date of birth cannot be in the future';
            dom.dobError.style.display = 'block';
            hasError = true;
        }
    }
    
    if (hasError) return;

    try {
        let res;
        const payload = { name, dob };
        
        if (state.isEditing) {
            res = await fetch(`/users/${state.editingUserId}`, {
                method: 'PUT',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            
            if (!res.ok) {
                const errorData = await res.json().catch(() => ({}));
                throw new Error(errorData.error || 'Failed to update user profile');
            }
            showToast('User profile updated successfully!', 'success');
        } else {
            res = await fetch('/users', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            
            if (!res.ok) {
                const errorData = await res.json().catch(() => ({}));
                throw new Error(errorData.error || 'Failed to create user');
            }
            showToast('User registered successfully!', 'success');
        }
        
        resetForm();
        loadUsers();
    } catch (err) {
        showToast(err.message, 'error');
    }
}

// INITIATE EDIT MODE
window.initiateEditUser = function(id, name, dob) {
    state.isEditing = true;
    state.editingUserId = id;
    
    dom.userIdInput.value = id;
    dom.nameInput.value = name;
    // Format date string to YYYY-MM-DD
    dom.dobInput.value = dob.substring(0, 10);
    
    dom.formTitle.textContent = 'Update User Profile';
    dom.formIcon.className = 'fa-solid fa-user-pen card-icon';
    dom.btnSubmitText.textContent = 'Save Changes';
    dom.btnCancel.style.display = 'inline-flex';
    
    // Scroll to form on mobile
    dom.form.scrollIntoView({ behavior: 'smooth' });
};

// DELETE USER HANDLER
window.handleDeleteUser = function(id, name) {
    openDeleteModal(id, name);
};

window.confirmDeleteUser = async function() {
    if (!state.deleteUserCandidateId) return;
    const id = state.deleteUserCandidateId;
    closeDeleteModal();

    try {
        const res = await fetch(`/users/${id}`, { method: 'DELETE' });
        if (!res.ok) throw new Error('Failed to delete user');

        showToast('User deleted successfully!', 'success');
        loadUsers();
    } catch (err) {
        showToast(err.message, 'error');
    }
};

function openDeleteModal(id, name) {
    state.deleteUserCandidateId = id;
    dom.deleteModalUser.textContent = escapeHTML(name);
    dom.deleteModalBackdrop.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
    dom.deleteModalConfirm.focus();
}

function closeDeleteModal() {
    state.deleteUserCandidateId = null;
    dom.deleteModalBackdrop.classList.add('hidden');
    document.body.style.overflow = '';
}

// RESET FORM STATE
function resetForm() {
    state.isEditing = false;
    state.editingUserId = null;
    
    dom.form.reset();
    dom.userIdInput.value = '';
    
    dom.formTitle.textContent = 'Register New User';
    dom.formIcon.className = 'fa-solid fa-user-plus card-icon';
    dom.btnSubmitText.textContent = 'Create User';
    dom.btnCancel.style.display = 'none';
    
    dom.nameError.style.display = 'none';
    dom.dobError.style.display = 'none';
}

// UPDATE PAGINATION UI
function updatePaginationUI() {
    dom.totalUsersBadge.textContent = `${state.totalUsers} User${state.totalUsers === 1 ? '' : 's'}`;
    
    dom.btnPrev.disabled = state.currentPage === 1;
    dom.btnNext.disabled = state.currentPage === state.totalPages || state.totalUsers === 0;
    
    dom.paginationInfo.textContent = `Page ${state.currentPage} of ${state.totalPages}`;
}

// HELPER: DATE FORMATTING
function formatDateString(dateStr) {
    try {
        const date = new Date(dateStr);
        return date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    } catch (e) {
        return dateStr;
    }
}

// LOADING STATES
function showTableLoading() {
    dom.tableBody.innerHTML = `
        <tr>
            <td colspan="5" class="table-loading">
                <div class="spinner"></div>
                <span>Syncing registry list...</span>
            </td>
        </tr>
    `;
}

function showTableError(message) {
    dom.tableBody.innerHTML = `
        <tr>
            <td colspan="5" class="table-empty" style="color: var(--danger);">
                <i class="fa-solid fa-triangle-exclamation" style="font-size: 2.5rem; margin-bottom: 1rem; display: block;"></i>
                <span>Database Sync Failed: ${escapeHTML(message)}</span>
            </td>
        </tr>
    `;
}

// HELPER: ESCAPE STRINGS
function escapeHTML(str) {
    return str
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#039;');
}

function escapeQuote(str) {
    return str.replace(/'/g, "\\'");
}
