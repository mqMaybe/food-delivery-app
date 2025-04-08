// Функция для отображения уведомлений через toastContainer
function showToast(message, type = 'error') {
    const toastContainer = document.getElementById('toastContainer');
    if (!toastContainer) return;

    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    toastContainer.appendChild(toast);

    setTimeout(() => {
        toast.remove();
    }, 3000);
}

// Функция для проверки авторизации
async function checkAuth(requiredRole = null) {
    try {
        const response = await fetch('/api/session');
        if (!response.ok) {
            throw new Error('Не удалось проверить сессию');
        }
        const sessionData = await response.json();
        if (!sessionData.user_id) {
            window.location.href = '/login';
            return null;
        }
        if (requiredRole && sessionData.role !== requiredRole) {
            if (sessionData.role === 'customer') {
                window.location.href = '/restaurants';
            } else if (sessionData.role === 'restaurant') {
                window.location.href = '/restaurant-orders';
            }
            return null;
        }
        return sessionData;
    } catch (error) {
        console.error('Failed to check auth:', error);
        showToast(error.message);
        window.location.href = '/login';
        return null;
    }
}

// Функция для выхода из системы
async function logout() {
    try {
        const response = await fetch('/api/logout', { method: 'POST' });
        if (response.ok) {
            window.location.href = '/login';
        } else {
            throw new Error('Не удалось выйти из системы');
        }
    } catch (error) {
        console.error('Failed to logout:', error);
        showToast(error.message);
    }
}