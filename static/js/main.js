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

async function searchRestaurants() {
    const searchInput = document.getElementById('search-input').value.trim();
    if (!searchInput) {
        alert('Пожалуйста, введите название ресторана для поиска');
        return;
    }

    // Перенаправляем на страницу ресторанов с параметром поиска
    window.location.href = `/restaurants?search=${encodeURIComponent(searchInput)}`;
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
        const csrfToken = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        if (!csrfToken) {
            throw new Error('CSRF-токен не найден');
        }

        const response = await fetch('/api/logout', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-CSRF-Token': csrfToken,
            },
        });

        const contentType = response.headers.get('Content-Type');
        let data = {};
        if (contentType && contentType.includes('application/json')) {
            data = await response.json();
        } else {
            const text = await response.text();
            console.error('Ответ сервера не является JSON:', text);
        }

        if (!response.ok) {
            throw new Error(data.error || 'Не удалось выйти из системы');
        }

        window.location.href = '/login';
    } catch (error) {
        console.error('Ошибка при выходе из системы:', error);
        displayError(error.message);
    }
}
