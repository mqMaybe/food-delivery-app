document.addEventListener('DOMContentLoaded', async () => {
    const session = await checkAuth('restaurant');
    if (!session) return;

    const addMenuBtn = document.getElementById('addMenuBtn');
    if (!addMenuBtn) {
        console.error('Button with id "addMenuBtn" not found');
        return;
    }

    addMenuBtn.addEventListener('click', async () => {
        const restaurantId = document.getElementById('menuRestaurantId').value;
        const name = document.getElementById('menuName').value;
        const price = document.getElementById('menuPrice').value;
        const description = document.getElementById('menuDescription').value;

        if (!restaurantId || !name || !price || !description) {
            displayError('addMenuResult', 'Пожалуйста, заполните все поля');
            return;
        }

        try {
            const response = await fetch('/api/menu', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    restaurant_id: parseInt(restaurantId),
                    name,
                    price: parseFloat(price),
                    description,
                }),
            });

            const data = await response.json();
            if (!response.ok) {
                throw new Error(data.error || 'Не удалось добавить блюдо');
            }

            alert('Блюдо успешно добавлено!');
            document.getElementById('menuRestaurantId').value = '';
            document.getElementById('menuName').value = '';
            document.getElementById('menuPrice').value = '';
            document.getElementById('menuDescription').value = '';
        } catch (error) {
            console.error('Failed to add menu item:', error);
            displayError('addMenuResult', error.message);
        }
    });
});