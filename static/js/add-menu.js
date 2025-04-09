// static/js/add-menu.js

// Загрузка списка ресторанов и обработка формы
document.addEventListener('DOMContentLoaded', async () => {
    // Загружаем список ресторанов
    try {
        const response = await fetch('/api/restaurants/user');
        console.log('Статус ответа:', response.status);
        console.log('Заголовки ответа:', response.headers.get('Content-Type'));

        const responseText = await response.text();
        console.log('Текст ответа:', responseText);

        if (!response.ok) {
            let errorData;
            try {
                errorData = JSON.parse(responseText);
            } catch (e) {
                throw new Error('Ответ сервера не является JSON: ' + responseText);
            }
            throw new Error(errorData.error || 'Не удалось загрузить рестораны');
        }

        const data = JSON.parse(responseText); // Пробуем разобрать как JSON
        const restaurants = data.restaurants;

        const restaurantSelect = document.getElementById('restaurant_id');
        restaurants.forEach(restaurant => {
            const option = document.createElement('option');
            option.value = restaurant.id;
            option.textContent = restaurant.name;
            restaurantSelect.appendChild(option);
        });
    } catch (error) {
        console.error('Ошибка при загрузке ресторанов:', error);
        displayError('addMenuError', error.message);
    }

    // Обработка формы
    const form = document.getElementById('addMenuForm');
    if (!form) {
        console.error('Форма для добавления блюда не найдена');
        return;
    }

    form.addEventListener('submit', async (event) => {
        event.preventDefault();

        const formData = new FormData(form);
        const menuItem = {
            restaurant_id: parseInt(formData.get('restaurant_id')),
            name: formData.get('name'),
            price: parseFloat(formData.get('price')),
            description: formData.get('description'),
            image_url: formData.get('image_url'),
        };

        try {
            const csrfToken = document.querySelector('meta[name="csrf-token"]').getAttribute('content');
            if (!csrfToken) {
                throw new Error('CSRF-токен не найден');
            }

            const response = await fetch('/api/menu', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'X-CSRF-Token': csrfToken,
                },
                body: JSON.stringify(menuItem),
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
                throw new Error(data.error || 'Не удалось добавить блюдо');
            }

            alert('Блюдо успешно добавлено!');
            form.reset();
        } catch (error) {
            console.error('Failed to add menu item:', error);
            displayError('addMenuError', error.message);
        }
    });
});

function displayError(elementId, message) {
    const element = document.getElementById(elementId);
    if (element) {
        element.innerHTML = `<p class="text-danger">${message}</p>`;
    }
}

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
        displayError('addMenuError', error.message);
    }
}