// Global App State
const App = {
    data: {
        venues: [],
        events: [],
        exhibitions: [],
        categories: [],
        filteredVenues: [],
        filteredEvents: []
    },
    state: {
        currentView: 'map',
        selectedVenue: null,
        selectedEvent: null,
        filters: {
            search: '',
            date: '',
            category: '',
            bikeRoute: false
        }
    },
    map: null,
    markers: null
};

// API Base URL
const API_BASE = '/api';

// Initialize App
document.addEventListener('DOMContentLoaded', async () => {
    console.log('🎨 Kulturelle Landpartie App initializing...');
    
    // Show loading
    showLoading();
    
    try {
        // Load data
        await loadData();
        
        // Initialize components
        initNavigation();
        initMap();
        initFilters();
        initFavorites();
        initSearch();
        
        // Apply initial filters
        applyFilters();
        
        // Hide loading
        hideLoading();
        
        console.log('✅ App initialized successfully');
    } catch (error) {
        console.error('❌ Failed to initialize app:', error);
        hideLoading();
        showError('Fehler beim Laden der Daten. Bitte versuchen Sie es später erneut.');
    }
});

// Load Data from API
async function loadData() {
    console.log('Loading data from API...');
    
    try {
        // Load venues
        const venuesResponse = await fetch(`${API_BASE}/venues`);
        const venuesData = await venuesResponse.json();
        App.data.venues = venuesData.venues || [];
        console.log(`✓ Loaded ${App.data.venues.length} venues`);
        
        // Load events
        const eventsResponse = await fetch(`${API_BASE}/events`);
        const eventsData = await eventsResponse.json();
        App.data.events = eventsData.events || [];
        console.log(`✓ Loaded ${App.data.events.length} events`);
        
        // Load exhibitions
        const exhibitionsResponse = await fetch(`${API_BASE}/exhibitions`);
        const exhibitionsData = await exhibitionsResponse.json();
        App.data.exhibitions = exhibitionsData.exhibitions || [];
        console.log(`✓ Loaded ${App.data.exhibitions.length} exhibitions`);
        
        // Load categories
        const categoriesResponse = await fetch(`${API_BASE}/categories`);
        const categoriesData = await categoriesResponse.json();
        App.data.categories = categoriesData.categories || [];
        console.log(`✓ Loaded ${App.data.categories.length} categories`);
        
        // Populate filter dropdowns
        populateFilters();
        
    } catch (error) {
        console.error('Error loading data:', error);
        throw error;
    }
}

// Populate Filter Dropdowns
function populateFilters() {
    // Populate date filter
    const dateFilter = document.getElementById('date-filter');
    const dates = [...new Set(App.data.events.map(e => e.date))].sort();
    
    dates.forEach(date => {
        const option = document.createElement('option');
        option.value = date;
        option.textContent = formatDate(date);
        dateFilter.appendChild(option);
    });
    
    // Populate category filter
    const categoryFilter = document.getElementById('category-filter');
    App.data.categories.forEach(cat => {
        const option = document.createElement('option');
        option.value = cat.name;
        option.textContent = `${cat.name} (${cat.count})`;
        categoryFilter.appendChild(option);
    });
}

// Initialize Navigation
function initNavigation() {
    const navButtons = document.querySelectorAll('.nav-btn');
    
    navButtons.forEach(btn => {
        btn.addEventListener('click', () => {
            const view = btn.dataset.view;
            switchView(view);
        });
    });
}

// Switch View
function switchView(view) {
    // Update state
    App.state.currentView = view;
    
    // Update nav buttons
    document.querySelectorAll('.nav-btn').forEach(btn => {
        btn.classList.toggle('active', btn.dataset.view === view);
    });
    
    // Update views
    document.querySelectorAll('.view').forEach(v => {
        v.classList.toggle('active', v.id === `${view}-view`);
    });
    
    // Load view-specific content
    if (view === 'calendar') {
        loadCalendar();
    } else if (view === 'favorites') {
        loadFavorites();
    } else if (view === 'map') {
        // Refresh map
        if (App.map) {
            setTimeout(() => App.map.invalidateSize(), 100);
        }
    }
}

// Apply Filters
function applyFilters() {
    const { search, date, category, bikeRoute } = App.state.filters;
    
    // Filter venues
    App.data.filteredVenues = App.data.venues.filter(venue => {
        // Search filter
        if (search && !venue.name.toLowerCase().includes(search.toLowerCase()) &&
            !venue.description.toLowerCase().includes(search.toLowerCase())) {
            return false;
        }
        
        // Bike route filter
        if (bikeRoute && !venue.bikeRoute) {
            return false;
        }
        
        return true;
    });
    
    // Filter events
    App.data.filteredEvents = App.data.events.filter(event => {
        // Date filter
        if (date && event.date !== date) {
            return false;
        }
        
        // Category filter
        if (category && event.category !== category) {
            return false;
        }
        
        // Search filter
        if (search && !event.title.toLowerCase().includes(search.toLowerCase()) &&
            !event.description.toLowerCase().includes(search.toLowerCase())) {
            return false;
        }
        
        return true;
    });
    
    // Update UI
    updateResults();
    updateMap();
}

// Update Results List
function updateResults() {
    const resultsList = document.getElementById('results-list');
    const resultsCount = document.getElementById('results-count');
    
    resultsList.innerHTML = '';
    
    // Combine venues and events
    const results = [
        ...App.data.filteredVenues.map(v => ({ type: 'venue', data: v })),
        ...App.data.filteredEvents.map(e => ({ type: 'event', data: e }))
    ];
    
    resultsCount.textContent = results.length;
    
    results.forEach(result => {
        const item = createResultItem(result);
        resultsList.appendChild(item);
    });
}

// Create Result Item
function createResultItem(result) {
    const div = document.createElement('div');
    div.className = 'result-item';
    
    if (result.type === 'venue') {
        const venue = result.data;
        div.innerHTML = `
            <h4><i class="fas fa-map-marker-alt"></i> ${venue.name}</h4>
            <p>${venue.address.city}</p>
            <div class="meta">
                <span><i class="fas fa-calendar"></i> ${venue.eventCount} Events</span>
                <span><i class="fas fa-palette"></i> ${venue.exhibitionCount} Ausstellungen</span>
            </div>
        `;
        div.addEventListener('click', () => showVenueDetails(venue.id));
    } else {
        const event = result.data;
        div.innerHTML = `
            <h4><i class="fas fa-calendar-day"></i> ${event.title}</h4>
            <p>${event.venueName || ''}</p>
            <div class="meta">
                <span><i class="fas fa-clock"></i> ${formatDate(event.date)} ${event.startTime || ''}</span>
                ${event.category ? `<span><i class="fas fa-tag"></i> ${event.category}</span>` : ''}
            </div>
        `;
        div.addEventListener('click', () => showEventDetails(event.id));
    }
    
    return div;
}

// Show Venue Details
async function showVenueDetails(venueId) {
    showLoading();
    
    try {
        const response = await fetch(`${API_BASE}/venues/${venueId}`);
        const venue = await response.json();
        
        const modal = document.getElementById('detail-modal');
        const content = document.getElementById('detail-content');
        
        content.innerHTML = `
            <h2>${venue.name}</h2>
            <p>${venue.description || ''}</p>
            
            <h3><i class="fas fa-map-marker-alt"></i> Adresse</h3>
            <p>${venue.address.street}<br>${venue.address.postalCode} ${venue.address.city}</p>
            
            ${venue.contact.phone ? `<p><i class="fas fa-phone"></i> ${venue.contact.phone}</p>` : ''}
            ${venue.contact.email ? `<p><i class="fas fa-envelope"></i> ${venue.contact.email}</p>` : ''}
            ${venue.contact.website ? `<p><i class="fas fa-globe"></i> <a href="${venue.contact.website}" target="_blank">${venue.contact.website}</a></p>` : ''}
            
            ${venue.events && venue.events.length > 0 ? `
                <h3><i class="fas fa-calendar"></i> Veranstaltungen (${venue.events.length})</h3>
                <div class="event-grid">
                    ${venue.events.map(e => `
                        <div class="event-card" onclick="showEventDetails('${e.id}')">
                            <div class="time">${e.startTime || ''}</div>
                            <h4>${e.title}</h4>
                            ${e.category ? `<span class="category">${e.category}</span>` : ''}
                        </div>
                    `).join('')}
                </div>
            ` : ''}
            
            ${venue.exhibitions && venue.exhibitions.length > 0 ? `
                <h3><i class="fas fa-palette"></i> Ausstellungen (${venue.exhibitions.length})</h3>
                <div class="event-grid">
                    ${venue.exhibitions.map(ex => `
                        <div class="event-card">
                            <h4>${ex.title}</h4>
                            ${ex.artist ? `<p>${ex.artist}</p>` : ''}
                            ${ex.category ? `<span class="category">${ex.category}</span>` : ''}
                        </div>
                    `).join('')}
                </div>
            ` : ''}
            
            <div class="popup-actions">
                <button class="btn-primary" onclick="centerMapOnVenue('${venue.id}')">
                    <i class="fas fa-map"></i> Auf Karte zeigen
                </button>
                <button class="btn-icon ${isFavorite(venue.id) ? 'active' : ''}" onclick="toggleFavorite('${venue.id}', 'venue')">
                    <i class="fas fa-heart"></i>
                </button>
            </div>
        `;
        
        modal.classList.add('active');
        hideLoading();
    } catch (error) {
        console.error('Error loading venue details:', error);
        hideLoading();
    }
}

// Show Event Details
async function showEventDetails(eventId) {
    showLoading();
    
    try {
        const response = await fetch(`${API_BASE}/events/${eventId}`);
        const event = await response.json();
        
        const modal = document.getElementById('detail-modal');
        const content = document.getElementById('detail-content');
        
        content.innerHTML = `
            <h2>${event.title}</h2>
            <p>${event.description || ''}</p>
            
            <h3><i class="fas fa-calendar"></i> Wann</h3>
            <p>${formatDate(event.date)} ${event.startTime ? `um ${event.startTime} Uhr` : ''}</p>
            
            ${event.venue ? `
                <h3><i class="fas fa-map-marker-alt"></i> Wo</h3>
                <p>${event.venue.name}<br>${event.venue.address.street}<br>${event.venue.address.postalCode} ${event.venue.address.city}</p>
            ` : ''}
            
            ${event.admission ? `<p><i class="fas fa-ticket-alt"></i> Eintritt: ${event.admission}</p>` : ''}
            ${event.category ? `<p><i class="fas fa-tag"></i> Kategorie: ${event.category}</p>` : ''}
            
            <div class="popup-actions">
                ${event.venue ? `
                    <button class="btn-primary" onclick="centerMapOnVenue('${event.venueId}')">
                        <i class="fas fa-map"></i> Auf Karte zeigen
                    </button>
                ` : ''}
                <button class="btn-icon ${isFavorite(eventId) ? 'active' : ''}" onclick="toggleFavorite('${eventId}', 'event')">
                    <i class="fas fa-heart"></i>
                </button>
            </div>
        `;
        
        modal.classList.add('active');
        hideLoading();
    } catch (error) {
        console.error('Error loading event details:', error);
        hideLoading();
    }
}

// Show Exhibition Details
async function showExhibitionDetails(exhibitionId) {
    showLoading();
    
    try {
        const response = await fetch(`${API_BASE}/exhibitions/${exhibitionId}`);
        const exhibition = await response.json();
        
        const modal = document.getElementById('detail-modal');
        const content = document.getElementById('detail-content');
        
        content.innerHTML = `
            <h2>${exhibition.title}</h2>
            ${exhibition.artist ? `<p style="font-style: italic; color: #666; margin-top: -0.5rem;"><i class="fas fa-palette"></i> ${exhibition.artist}</p>` : ''}
            ${exhibition.description ? `<p>${exhibition.description}</p>` : ''}
            
            ${exhibition.venue ? `
                <h3><i class="fas fa-map-marker-alt"></i> Wo</h3>
                <p>${exhibition.venue.name}<br>${exhibition.venue.address.street}<br>${exhibition.venue.address.postalCode} ${exhibition.venue.address.city}</p>
            ` : ''}
            
            ${exhibition.category ? `<p><i class="fas fa-tag"></i> Kategorie: ${exhibition.category}</p>` : ''}
            
            <div class="popup-actions">
                ${exhibition.venue ? `
                    <button class="btn-primary" onclick="centerMapOnVenue('${exhibition.venueId}')">
                        <i class="fas fa-map"></i> Auf Karte zeigen
                    </button>
                ` : ''}
                <button class="btn-icon ${isFavorite(exhibitionId) ? 'active' : ''}" onclick="toggleFavorite('${exhibitionId}', 'exhibition')">
                    <i class="fas fa-heart"></i>
                </button>
            </div>
        `;
        
        modal.classList.add('active');
        hideLoading();
    } catch (error) {
        console.error('Error loading exhibition details:', error);
        hideLoading();
    }
}

// Close Modal
document.querySelector('.modal-close')?.addEventListener('click', () => {
    document.getElementById('detail-modal').classList.remove('active');
});

document.getElementById('detail-modal')?.addEventListener('click', (e) => {
    if (e.target.id === 'detail-modal') {
        document.getElementById('detail-modal').classList.remove('active');
    }
});

// Utility Functions
function formatDate(dateStr) {
    const date = new Date(dateStr);
    const days = ['Sonntag', 'Montag', 'Dienstag', 'Mittwoch', 'Donnerstag', 'Freitag', 'Samstag'];
    const months = ['Jan', 'Feb', 'Mär', 'Apr', 'Mai', 'Jun', 'Jul', 'Aug', 'Sep', 'Okt', 'Nov', 'Dez'];
    
    return `${days[date.getDay()]}, ${date.getDate()}. ${months[date.getMonth()]}`;
}

function showLoading() {
    document.getElementById('loading').classList.add('active');
}

function hideLoading() {
    document.getElementById('loading').classList.remove('active');
}

function showError(message) {
    alert(message); // Simple error handling - could be improved with a toast notification
}

// Export for use in other modules
window.App = App;
window.showVenueDetails = showVenueDetails;
window.showEventDetails = showEventDetails;
window.showExhibitionDetails = showExhibitionDetails;
