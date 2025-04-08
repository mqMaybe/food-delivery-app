// static/js/restaurants.js

// Функция для отображения ошибки
function displayError(elementId, message) {
    const element = document.getElementById(elementId);
    if (element) {
        element.innerHTML = `<p class="text-danger">${message}</p>`;
    }
}

// Функция для выхода из системы
async function logout() {
    try {
        const response = await fetch('/api/logout', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
        });

        const data = await response.json();
        if (!response.ok) {
            throw new Error(data.error || 'Не удалось выйти из системы');
        }

        window.location.href = '/login';
    } catch (error) {
        console.error('Ошибка при выходе из системы:', error);
        displayError('restaurantList', error.message);
    }
}

// Загрузка ресторанов при загрузке страницы
document.addEventListener('DOMContentLoaded', async () => {
    console.log('restaurants.js загружен'); // Отладка

    const restaurantList = document.getElementById('restaurantList');
    if (!restaurantList) {
        console.error('Элемент списка ресторанов не найден');
        return;
    }

    const cuisineType = document.getElementById('cuisineType');
    const deliveryTime = document.getElementById('deliveryTime');
    const rating = document.getElementById('rating');

    const loadRestaurants = async () => {
        try {
            console.log('Загрузка ресторанов...'); // Отладка
            const params = new URLSearchParams();
            if (cuisineType && cuisineType.value !== 'all') {
                params.append('cuisine_type', cuisineType.value);
            }
            if (deliveryTime && deliveryTime.value !== 'all') {
                params.append('delivery_time', deliveryTime.value);
            }
            if (rating && rating.value !== 'all') {
                params.append('rating', rating.value);
            }

            console.log('URL запроса:', `/api/restaurants?${params.toString()}`); // Отладка
            const response = await fetch(`/api/restaurants?${params.toString()}`);
            console.log('Статус ответа:', response.status); // Отладка

            if (!response.ok) {
                const errorData = await response.json();
                console.error('Ошибка в ответе:', errorData); // Отладка
                throw new Error(errorData.error || 'Не удалось загрузить рестораны');
            }

            const restaurants = await response.json();
            console.log('Рестораны:', restaurants); // Отладка

            restaurantList.innerHTML = ''; // Очищаем список
            if (restaurants.length === 0) {
                restaurantList.innerHTML = '<p>Рестораны не найдены.</p>';
                return;
            }

            restaurants.forEach(restaurant => {
                const col = document.createElement('div');
                col.className = 'col-md-4';
                col.innerHTML = `
                    <div class="restaurant-card">
                        <img src="/static/images/restaurant-placeholder.jpg" alt="${restaurant.name}" class="restaurant-image">
                        <div class="restaurant-name">${restaurant.name}</div>
                        <div class="restaurant-rating">Рейтинг: ${restaurant.rating.toFixed(1)}</div>
                        <a href="/menu?restaurant_id=${restaurant.id}" class="btn view-details-btn">Подробнее</a>
                    </div>
                `;
                restaurantList.appendChild(col);
            });
        } catch (error) {
            console.error('Ошибка при загрузке ресторанов:', error);
            displayError('restaurantList', error.message);
        }
    };

    // Загружаем рестораны при загрузке страницы
    await loadRestaurants();

    // Добавляем обработчики для фильтров
    if (cuisineType && deliveryTime && rating) {
        cuisineType.addEventListener('change', loadRestaurants);
        deliveryTime.addEventListener('change', loadRestaurants);
        rating.addEventListener('change', loadRestaurants);
    } else {
        console.error('Один или несколько элементов фильтров отсутствуют:', {
            cuisineType: !!cuisineType,
            deliveryTime: !!deliveryTime,
            rating: !!rating,
        });
    }
});