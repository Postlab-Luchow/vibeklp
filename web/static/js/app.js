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
        showError('Fehler beim Laden der Daten. Bitte versuchen Sie es später erneut.', {
            retry: 'location.reload()'
        });
    }
});

// Load Data from API with retry logic
async function loadData(retryCount = 0) {
    console.log('Loading data from API...');
    
    const maxRetries = 2;
    const failedEndpoints = [];
    
    try {
        // Load venues with retry
        try {
            const venuesResponse = await fetchWithRetry(`${API_BASE}/venues`);
            const venuesData = await venuesResponse.json();
            App.data.venues = venuesData.venues || [];
            console.log(`✓ Loaded ${App.data.venues.length} venues`);
        } catch (error) {
            console.error('Failed to load venues:', error);
            failedEndpoints.push('Orte');
        }
        
        // Load events with retry
        try {
            const eventsResponse = await fetchWithRetry(`${API_BASE}/events`);
            const eventsData = await eventsResponse.json();
            App.data.events = eventsData.events || [];
            console.log(`✓ Loaded ${App.data.events.length} events`);
        } catch (error) {
            console.error('Failed to load events:', error);
            failedEndpoints.push('Events');
        }
        
        // Load exhibitions with retry
        try {
            const exhibitionsResponse = await fetchWithRetry(`${API_BASE}/exhibitions`);
            const exhibitionsData = await exhibitionsResponse.json();
            App.data.exhibitions = exhibitionsData.exhibitions || [];
            console.log(`✓ Loaded ${App.data.exhibitions.length} exhibitions`);
        } catch (error) {
            console.error('Failed to load exhibitions:', error);
            failedEndpoints.push('Ausstellungen');
        }
        
        // Load categories with retry
        try {
            const categoriesResponse = await fetchWithRetry(`${API_BASE}/categories`);
            const categoriesData = await categoriesResponse.json();
            App.data.categories = categoriesData.categories || [];
            console.log(`✓ Loaded ${App.data.categories.length} categories`);
        } catch (error) {
            console.error('Failed to load categories:', error);
            failedEndpoints.push('Kategorien');
        }
        
        // Show warning if some endpoints failed but app can still work
        if (failedEndpoints.length > 0 && failedEndpoints.length < 4) {
            const message = `Einige Daten konnten nicht geladen werden: ${failedEndpoints.join(', ')}`;
            showError(message, {
                retry: retryCount < maxRetries ? 'loadData(' + (retryCount + 1) + ')' : null
            });
        }
        
        // Populate filter dropdowns
        populateFilters();
        
        // Throw error if all endpoints failed
        if (failedEndpoints.length === 4) {
            throw new Error('Alle Datenquellen konnten nicht geladen werden');
        }
        
    } catch (error) {
        console.error('Error loading data:', error);
        throw error;
    }
}

// Fetch with retry logic
async function fetchWithRetry(url, options = {}, maxRetries = 2) {
    let lastError;
    
    for (let i = 0; i <= maxRetries; i++) {
        try {
            const response = await fetch(url, options);
            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
            return response;
        } catch (error) {
            lastError = error;
            if (i < maxRetries) {
                console.warn(`Retry ${i + 1}/${maxRetries} for ${url}...`);
                await new Promise(resolve => setTimeout(resolve, 1000 * (i + 1)));
            }
        }
    }
    
    throw lastError;
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
    const searchLower = search ? search.toLowerCase() : '';
    
    // Build venue → events / exhibitions lookups from the venueId on each
    // child item. (The crawler does not populate venue.eventIds /
    // venue.exhibitionIds, so we cannot rely on those.)
    const eventsByVenue = new Map();
    for (const e of App.data.events) {
        if (!e.venueId) continue;
        const list = eventsByVenue.get(e.venueId);
        if (list) list.push(e); else eventsByVenue.set(e.venueId, [e]);
    }
    const exhibitionsByVenue = new Map();
    for (const ex of App.data.exhibitions) {
        if (!ex.venueId) continue;
        const list = exhibitionsByVenue.get(ex.venueId);
        if (list) list.push(ex); else exhibitionsByVenue.set(ex.venueId, [ex]);
    }

    // Filter venues (search includes venue data + events + exhibitions)
    App.data.filteredVenues = App.data.venues.filter(venue => {
        // Bike route filter
        if (bikeRoute && !venue.bikeRoute) {
            return false;
        }

        const venueEvents = eventsByVenue.get(venue.id) || [];
        const venueExhibitions = exhibitionsByVenue.get(venue.id) || [];
        
        // Date filter - venue must have event on this date
        if (date) {
            const hasEventOnDate = venueEvents.some(e => e.date === date);
            if (!hasEventOnDate) return false;
        }
        
        // Category filter - venue must offer this facility (Café, WC, …).
        // If a date filter is also active and the category has date restrictions,
        // the category must be available on that specific date.
        if (category) {
            const match = (venue.categories || []).find(c => c.name === category);
            if (!match) return false;
            if (date && match.dates && match.dates.length > 0 && !match.dates.includes(date)) {
                return false;
            }
        }
        
        // Search filter - check venue + events + exhibitions
        if (searchLower) {
            // Check venue fields (with null-safety)
            const venueMatches = 
                venue.name.toLowerCase().includes(searchLower) ||
                (venue.description && venue.description.toLowerCase().includes(searchLower)) ||
                (venue.address && venue.address.city && venue.address.city.toLowerCase().includes(searchLower)) ||
                (venue.address && venue.address.street && venue.address.street.toLowerCase().includes(searchLower));
            
            if (venueMatches) return true;
            
            // Check events at this venue
            const eventMatches = venueEvents.some(e => 
                e.title.toLowerCase().includes(searchLower) ||
                (e.description && e.description.toLowerCase().includes(searchLower)) ||
                (e.category && e.category.toLowerCase().includes(searchLower))
            );
            
            if (eventMatches) return true;
            
            // Check exhibitions at this venue
            const exhibitionMatches = venueExhibitions.some(ex => 
                ex.title.toLowerCase().includes(searchLower) ||
                (ex.description && ex.description.toLowerCase().includes(searchLower)) ||
                (ex.artist && ex.artist.toLowerCase().includes(searchLower))
            );
            
            if (!exhibitionMatches) return false;
        }
        
        return true;
    });
    
    // Store filtered events for map display (events at filtered venues)
    const filteredVenueIds = new Set(App.data.filteredVenues.map(v => v.id));
    App.data.filteredEvents = App.data.events.filter(e => filteredVenueIds.has(e.venueId));
    
    // Update UI
    updateResults();
    updateMap();
    if (App.state.currentView === 'calendar') {
        loadCalendar();
    }
}

// Update Results List - Show only venues
function updateResults() {
    const resultsList = document.getElementById('results-list');
    const resultsCount = document.getElementById('results-count');
    
    resultsList.innerHTML = '';
    resultsCount.textContent = App.data.filteredVenues.length;
    
    // Show empty state if no results
    if (App.data.filteredVenues.length === 0) {
        resultsList.innerHTML = `
            <div class="results-empty">
                <i class="fas fa-search"></i>
                <p>Keine Ergebnisse gefunden</p>
                <p class="hint">Versuchen Sie andere Suchbegriffe oder Filter</p>
            </div>
        `;
        return;
    }
    
    // Show only venues in results list
    App.data.filteredVenues.forEach(venue => {
        const item = createResultItem(venue);
        resultsList.appendChild(item);
    });
}

// Create Result Item for Venue
function createResultItem(venue) {
    const div = document.createElement('div');
    div.className = 'result-item';
    
    // Calculate matching events/exhibitions for search highlight
    const { search } = App.state.filters;
    let matchInfo = '';
    
    if (search && search.length >= 2) {
        const searchLower = search.toLowerCase();
        const venueEvents = App.data.events.filter(e => e.venueId === venue.id);
        const venueExhibitions = App.data.exhibitions.filter(ex => ex.venueId === venue.id);
        
        // Count matching events
        const matchingEvents = venueEvents.filter(e => 
            e.title.toLowerCase().includes(searchLower) ||
            (e.description && e.description.toLowerCase().includes(searchLower))
        ).length;
        
        // Count matching exhibitions
        const matchingExhibitions = venueExhibitions.filter(ex => 
            ex.title.toLowerCase().includes(searchLower) ||
            (ex.description && ex.description.toLowerCase().includes(searchLower)) ||
            (ex.artist && ex.artist.toLowerCase().includes(searchLower))
        ).length;
        
        // Build match info string
        const matches = [];
        if (matchingEvents > 0) matches.push(`${matchingEvents} Event${matchingEvents > 1 ? 's' : ''}`);
        if (matchingExhibitions > 0) matches.push(`${matchingExhibitions} Ausstellung${matchingExhibitions > 1 ? 'en' : ''}`);
        
        if (matches.length > 0) {
            matchInfo = `<span class="match-badge"><i class="fas fa-search"></i> ${matches.join(', ')}</span>`;
        }
    }
    
    div.innerHTML = `
        <h4><i class="fas fa-map-marker-alt"></i> ${venue.name}</h4>
        <p>${venue.address.city}</p>
        <div class="meta">
            <span><i class="fas fa-calendar"></i> ${venue.eventCount || 0} Events</span>
            <span><i class="fas fa-palette"></i> ${venue.exhibitionCount || 0} Ausstellungen</span>
            ${matchInfo}
        </div>
    `;
    
    div.addEventListener('click', () => showVenueDetails(venue.id));
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

function showError(message, options = {}) {
    const container = document.getElementById('error-container');
    if (!container) {
        console.error('Error container not found:', message);
        return;
    }
    
    const errorId = 'error-' + Date.now();
    const errorDiv = document.createElement('div');
    errorDiv.id = errorId;
    errorDiv.className = 'error-message';
    errorDiv.innerHTML = `
        <i class="fas fa-exclamation-circle"></i>
        <span>${escapeHtml(message)}</span>
        ${options.retry ? `<button class="retry-btn" onclick="${options.retry}">Erneut versuchen</button>` : ''}
        <button class="close-btn" onclick="dismissError('${errorId}')" aria-label="Fehler schließen">
            <i class="fas fa-times"></i>
        </button>
    `;
    
    container.appendChild(errorDiv);
    
    // Auto-dismiss after 10 seconds unless it's a retryable error
    if (!options.retry) {
        setTimeout(() => dismissError(errorId), 10000);
    }
}

function dismissError(errorId) {
    const errorDiv = document.getElementById(errorId);
    if (errorDiv) {
        errorDiv.style.opacity = '0';
        errorDiv.style.transform = 'translateY(-20px)';
        setTimeout(() => errorDiv.remove(), 300);
    }
}

function showSuccess(message) {
    const container = document.getElementById('error-container');
    if (!container) {
        console.log('Success:', message);
        return;
    }
    
    const successId = 'success-' + Date.now();
    const successDiv = document.createElement('div');
    successDiv.id = successId;
    successDiv.className = 'success-message';
    successDiv.innerHTML = `
        <i class="fas fa-check-circle"></i>
        <span>${escapeHtml(message)}</span>
        <button class="close-btn" onclick="dismissError('${successId}')" aria-label="Nachricht schließen" style="color: #155724;">
            <i class="fas fa-times"></i>
        </button>
    `;
    
    container.appendChild(successDiv);
    
    // Auto-dismiss after 5 seconds
    setTimeout(() => dismissError(successId), 5000);
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// Export for use in other modules
window.App = App;
window.showVenueDetails = showVenueDetails;
window.showEventDetails = showEventDetails;
window.showExhibitionDetails = showExhibitionDetails;
window.showError = showError;
window.showSuccess = showSuccess;
window.dismissError = dismissError;
