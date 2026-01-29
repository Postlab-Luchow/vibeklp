// Favorites Module
const FAVORITES_KEY = 'klp_favorites';

// Initialize Favorites
function initFavorites() {
    console.log('Initializing favorites...');
    updateFavoritesCount();
    console.log('✓ Favorites initialized');
}

// Get Favorites from LocalStorage
function getFavorites() {
    const stored = localStorage.getItem(FAVORITES_KEY);
    return stored ? JSON.parse(stored) : { venues: [], events: [], exhibitions: [] };
}

// Save Favorites to LocalStorage
function saveFavorites(favorites) {
    localStorage.setItem(FAVORITES_KEY, JSON.stringify(favorites));
    updateFavoritesCount();
}

// Check if Item is Favorite
function isFavorite(id) {
    const favorites = getFavorites();
    return favorites.venues.includes(id) || 
           favorites.events.includes(id) || 
           favorites.exhibitions.includes(id);
}

// Toggle Favorite
function toggleFavorite(id, type) {
    const favorites = getFavorites();
    const key = type + 's'; // venues, events, exhibitions
    
    if (!favorites[key]) {
        favorites[key] = [];
    }
    
    const index = favorites[key].indexOf(id);
    
    if (index > -1) {
        // Remove from favorites
        favorites[key].splice(index, 1);
        console.log(`Removed ${type} ${id} from favorites`);
    } else {
        // Add to favorites
        favorites[key].push(id);
        console.log(`Added ${type} ${id} to favorites`);
    }
    
    saveFavorites(favorites);
    
    // Update UI if on favorites view
    if (App.state.currentView === 'favorites') {
        loadFavorites();
    }
    
    // Update button states
    updateFavoriteButtons(id);
}

// Update Favorites Count
function updateFavoritesCount() {
    const favorites = getFavorites();
    const total = (favorites.venues?.length || 0) + 
                  (favorites.events?.length || 0) + 
                  (favorites.exhibitions?.length || 0);
    
    const badge = document.getElementById('favorites-count');
    if (badge) {
        badge.textContent = total;
        badge.style.display = total > 0 ? 'inline-block' : 'none';
    }
}

// Update Favorite Buttons
function updateFavoriteButtons(id) {
    const isFav = isFavorite(id);
    document.querySelectorAll(`[onclick*="'${id}'"]`).forEach(btn => {
        if (btn.classList.contains('btn-icon')) {
            btn.classList.toggle('active', isFav);
        }
    });
}

// Load Favorites View
function loadFavorites() {
    console.log('Loading favorites...');
    
    const favorites = getFavorites();
    const favoritesDiv = document.getElementById('favorites-list');
    favoritesDiv.innerHTML = '';
    
    const totalFavorites = (favorites.venues?.length || 0) + 
                          (favorites.events?.length || 0) + 
                          (favorites.exhibitions?.length || 0);
    
    if (totalFavorites === 0) {
        favoritesDiv.innerHTML = `
            <div class="favorites-empty">
                <i class="fas fa-heart-broken"></i>
                <h3>Keine Favoriten</h3>
                <p>Fügen Sie Orte und Events zu Ihren Favoriten hinzu, um sie hier zu sehen.</p>
            </div>
        `;
        return;
    }
    
    // Favorite Venues
    if (favorites.venues && favorites.venues.length > 0) {
        const venuesSection = document.createElement('div');
        venuesSection.className = 'calendar-day';
        venuesSection.innerHTML = `
            <h3>
                <span><i class="fas fa-map-marker-alt"></i> Orte</span>
                <span class="badge">${favorites.venues.length}</span>
            </h3>
            <div class="event-grid" id="favorite-venues"></div>
        `;
        favoritesDiv.appendChild(venuesSection);
        
        const venuesGrid = venuesSection.querySelector('#favorite-venues');
        favorites.venues.forEach(venueId => {
            const venue = App.data.venues.find(v => v.id === venueId);
            if (venue) {
                const card = createFavoriteVenueCard(venue);
                venuesGrid.appendChild(card);
            }
        });
    }
    
    // Favorite Events
    if (favorites.events && favorites.events.length > 0) {
        const eventsSection = document.createElement('div');
        eventsSection.className = 'calendar-day';
        eventsSection.innerHTML = `
            <h3>
                <span><i class="fas fa-calendar"></i> Veranstaltungen</span>
                <span class="badge">${favorites.events.length}</span>
            </h3>
            <div class="event-grid" id="favorite-events"></div>
        `;
        favoritesDiv.appendChild(eventsSection);
        
        const eventsGrid = eventsSection.querySelector('#favorite-events');
        favorites.events.forEach(eventId => {
            const event = App.data.events.find(e => e.id === eventId);
            if (event) {
                const card = createEventCard(event);
                eventsGrid.appendChild(card);
            }
        });
    }
    
    // Favorite Exhibitions
    if (favorites.exhibitions && favorites.exhibitions.length > 0) {
        const exhibitionsSection = document.createElement('div');
        exhibitionsSection.className = 'calendar-day';
        exhibitionsSection.innerHTML = `
            <h3>
                <span><i class="fas fa-palette"></i> Ausstellungen</span>
                <span class="badge">${favorites.exhibitions.length}</span>
            </h3>
            <div class="event-grid" id="favorite-exhibitions"></div>
        `;
        favoritesDiv.appendChild(exhibitionsSection);
        
        const exhibitionsGrid = exhibitionsSection.querySelector('#favorite-exhibitions');
        favorites.exhibitions.forEach(exhibitionId => {
            const exhibition = App.data.exhibitions.find(ex => ex.id === exhibitionId);
            if (exhibition) {
                const card = createFavoriteExhibitionCard(exhibition);
                exhibitionsGrid.appendChild(card);
            }
        });
    }
    
    console.log('✓ Favorites loaded');
}

// Create Favorite Venue Card
function createFavoriteVenueCard(venue) {
    const div = document.createElement('div');
    div.className = 'event-card';
    
    div.innerHTML = `
        <h4>${venue.name}</h4>
        <div class="venue">
            <i class="fas fa-map-marker-alt"></i> ${venue.address.city}
        </div>
        <div class="meta" style="margin-top: 0.5rem; font-size: 0.85rem;">
            <span><i class="fas fa-calendar"></i> ${venue.eventCount} Events</span>
            <span><i class="fas fa-palette"></i> ${venue.exhibitionCount} Ausstellungen</span>
        </div>
        <div class="popup-actions" style="margin-top: 1rem;">
            <button class="btn-icon active" onclick="toggleFavorite('${venue.id}', 'venue'); event.stopPropagation();" title="Aus Favoriten entfernen">
                <i class="fas fa-heart"></i>
            </button>
            <button class="btn-icon" onclick="centerMapOnVenue('${venue.id}'); event.stopPropagation();" title="Auf Karte zeigen">
                <i class="fas fa-map"></i>
            </button>
        </div>
    `;
    
    div.addEventListener('click', () => showVenueDetails(venue.id));
    
    return div;
}

// Create Favorite Exhibition Card
function createFavoriteExhibitionCard(exhibition) {
    const div = document.createElement('div');
    div.className = 'event-card';
    
    div.innerHTML = `
        <h4>${exhibition.title}</h4>
        ${exhibition.artist ? `<div class="venue"><i class="fas fa-user"></i> ${exhibition.artist}</div>` : ''}
        <div class="venue">
            <i class="fas fa-map-marker-alt"></i> ${exhibition.venueName || 'Ort nicht angegeben'}
        </div>
        ${exhibition.category ? `<span class="category">${exhibition.category}</span>` : ''}
        <div class="popup-actions" style="margin-top: 1rem;">
            <button class="btn-icon active" onclick="toggleFavorite('${exhibition.id}', 'exhibition'); event.stopPropagation();" title="Aus Favoriten entfernen">
                <i class="fas fa-heart"></i>
            </button>
        </div>
    `;
    
    return div;
}

// Export Favorites
function exportFavorites() {
    const favorites = getFavorites();
    const dataStr = JSON.stringify(favorites, null, 2);
    const dataBlob = new Blob([dataStr], { type: 'application/json' });
    
    const url = URL.createObjectURL(dataBlob);
    const link = document.createElement('a');
    link.href = url;
    link.download = 'klp-favorites.json';
    link.click();
    
    URL.revokeObjectURL(url);
}

// Import Favorites
function importFavorites(file) {
    const reader = new FileReader();
    reader.onload = (e) => {
        try {
            const favorites = JSON.parse(e.target.result);
            saveFavorites(favorites);
            loadFavorites();
            alert('Favoriten erfolgreich importiert!');
        } catch (error) {
            console.error('Error importing favorites:', error);
            alert('Fehler beim Importieren der Favoriten.');
        }
    };
    reader.readAsText(file);
}

// Export functions
window.initFavorites = initFavorites;
window.isFavorite = isFavorite;
window.toggleFavorite = toggleFavorite;
window.loadFavorites = loadFavorites;
window.exportFavorites = exportFavorites;
window.importFavorites = importFavorites;
