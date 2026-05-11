// Filters Module
function initFilters() {
    console.log('Initializing filters...');
    
    // Date filter
    document.getElementById('date-filter').addEventListener('change', (e) => {
        App.state.filters.date = e.target.value;
        applyFilters();
    });
    
    // Venue-amenity filter
    document.getElementById('category-filter').addEventListener('change', (e) => {
        App.state.filters.category = e.target.value;
        applyFilters();
    });

    // Event-category filter
    document.getElementById('event-category-filter')?.addEventListener('change', (e) => {
        App.state.filters.eventCategory = e.target.value;
        applyFilters();
    });
    
    // Bike route filter
    document.getElementById('bike-route-filter').addEventListener('change', (e) => {
        App.state.filters.bikeRoute = e.target.checked;
        applyFilters();
    });
    
    // Reset filters
    document.getElementById('reset-filters').addEventListener('click', resetFilters);
    
    console.log('✓ Filters initialized');
}

// Initialize Search
function initSearch() {
    const searchInput = document.getElementById('search-input');
    const searchBtn = document.getElementById('search-btn');
    
    // Debounce search input
    let searchTimeout;
    searchInput.addEventListener('input', (e) => {
        clearTimeout(searchTimeout);
        searchTimeout = setTimeout(() => {
            App.state.filters.search = e.target.value;
            applyFilters();
        }, 300);
    });
    
    // Search button
    searchBtn.addEventListener('click', () => {
        App.state.filters.search = searchInput.value;
        applyFilters();
    });
    
    // Enter key
    searchInput.addEventListener('keypress', (e) => {
        if (e.key === 'Enter') {
            App.state.filters.search = searchInput.value;
            applyFilters();
        }
    });
    
    console.log('✓ Search initialized');
}

// Reset Filters
function resetFilters() {
    // Reset state
    App.state.filters = {
        search: '',
        date: '',
        category: '',
        eventCategory: '',
        bikeRoute: false
    };

    // Reset UI
    document.getElementById('search-input').value = '';
    document.getElementById('date-filter').value = '';
    document.getElementById('category-filter').value = '';
    const ecFilter = document.getElementById('event-category-filter');
    if (ecFilter) ecFilter.value = '';
    document.getElementById('bike-route-filter').checked = false;

    // Apply filters
    applyFilters();
}

// Export functions
window.initFilters = initFilters;
window.initSearch = initSearch;
window.resetFilters = resetFilters;
