// static/js/register.js

document.addEventListener('DOMContentLoaded', () => {
    console.log('register.js загружен'); // Отладка

    const roleSelect = document.getElementById('role');
    const cuisineTypeSection = document.getElementById('cuisineTypeSection');
    const registerForm = document.getElementById('registerForm');
    const registerResult = document.getElementById('registerResult');
    const registerBtn = document.getElementById('registerBtn');

    if (!roleSelect || !cuisineTypeSection || !registerForm || !registerResult || !registerBtn) {
        console.error('Необходимые элементы не найдены');
        return;
    }

    // Показываем/скрываем поле cuisineType в зависимости от роли
    roleSelect.addEventListener('change', () => {
        if (roleSelect.value === 'restaurant') {
            cuisineTypeSection.style.display = 'block';
        } else {
            cuisineTypeSection.style.display = 'none';
        }
    });

    // Обработка клика по кнопке
    registerBtn.addEventListener('click', async (e) => {
        e.preventDefault(); // Предотвращаем отправку формы

        const name = document.getElementById('name').value;
        const email = document.getElementById('email').value;
        const password = document.getElementById('password').value;
        const role = roleSelect.value;
        const cuisineType = role === 'restaurant' ? document.getElementById('cuisineType').value : '';

        if (!name || !email || !password || !role) {
            registerResult.textContent = 'Пожалуйста, заполните все обязательные поля';
            return;
        }

        if (role === 'restaurant' && !cuisineType) {
            registerResult.textContent = 'Пожалуйста, выберите тип кухни';
            return;
        }

        // Получаем CSRF-токен из мета-тега
        const csrfToken = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
        if (!csrfToken) {
            console.error('CSRF-токен не найден');
            registerResult.textContent = 'Ошибка: CSRF-токен не найден';
            return;
        }

        try {
            const response = await fetch('/api/register', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken
                },
                body: JSON.stringify({ name, email, password, role, cuisine_type: cuisineType }),
            });

            // Проверяем, что сервер вернул JSON
            const contentType = response.headers.get('Content-Type');
            if (!contentType || !contentType.includes('application/json')) {
                const text = await response.text();
                console.error('Unexpected response:', text);
                throw new Error('Сервер вернул не JSON: ' + text);
            }

            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Не удалось зарегистрироваться');
            }

            alert('Регистрация прошла успешно! Пожалуйста, войдите.');
            window.location.href = '/login';
        } catch (error) {
            console.error('Ошибка при регистрации:', error);
            registerResult.textContent = error.message;
        }
    });
});