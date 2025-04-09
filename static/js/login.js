// static/js/login.js

document.addEventListener('DOMContentLoaded', () => {
const loginForm = document.getElementById('loginForm');
    const loginResult = document.getElementById('loginResult');
    const loginBtn = document.getElementById('loginBtn');

    // Проверяем, какие элементы не найдены
    if (!loginForm || !loginResult || !loginBtn) {
        const missingElements = [];
        if (!loginForm) missingElements.push('loginForm');
        if (!loginResult) missingElements.push('loginResult');
        if (!loginBtn) missingElements.push('loginBtn');
        console.error('Не найдены следующие элементы:', missingElements.join(', '));
        return;
    }

    loginBtn.addEventListener('click', async (e) => {
        e.preventDefault();

        const email = document.getElementById('loginEmail').value.trim();
        const password = document.getElementById('loginPassword').value.trim();

        console.log('Отправляемые данные:', { email, password });

        if (!email || !password) {
            loginResult.textContent = 'Пожалуйста, заполните все поля';
            return;
        }

        const csrfToken = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        if (!csrfToken) {
            console.error('CSRF-токен не найден');
            registerResult.textContent = 'Ошибка: CSRF-токен не найден';
            return;
        }

        try {
            const response = await fetch('/api/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken
                },
                body: JSON.stringify({ email, password }),
            });

            const contentType = response.headers.get('Content-Type');
            if (!contentType || !contentType.includes('application/json')) {
                const text = await response.text();
                console.error('Unexpected response:', text);
                throw new Error('Сервер вернул не JSON: ' + text);
            }

            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Не удалось войти');
            }
            
            // Перенаправление в зависимости от роли
            if (data.role === 'restaurant') {
                window.location.href = '/restaurant-orders';
            } else if (data.role === 'customer') {
                window.location.href = '/homepage';
            } else if (data.role === 'rider') {
                window.location.href = '/';
            } else {
                console.error('Неизвестная роль:', data.role);
                window.location.href = '/';
            }

            alert('Вход выполнен успешно!');
        } catch (error) {
            console.error('Ошибка при входе:', error);
            loginResult.textContent = error.message;
        }
    });

    function showToast(message) {
        const toastContainer = document.getElementById('toastContainer');
        const toast = document.createElement('div');
        toast.classList.add('toast');
        toast.textContent = message;

        toastContainer.appendChild(toast);

        setTimeout(() => {
            toast.classList.add('show');
        }, 100);

        setTimeout(() => {
            toast.classList.remove('show');
            setTimeout(() => {
                toast.remove();
            }, 300);
        }, 3000);
    }
});