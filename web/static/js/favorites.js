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
            <div class="text-center py-20 px-6 text-muted">
                <i class="fas fa-heart-broken text-4xl opacity-30 block mb-4"></i>
                <h3 class="text-base font-medium text-ink">Keine Favoriten</h3>
                <p class="text-sm mt-1 max-w-sm mx-auto">Fügen Sie Orte und Events zu Ihren Favoriten hinzu, um sie hier zu sehen.</p>
            </div>
        `;
        return;
    }

    const renderSection = (label, iconClass, count, gridId) => {
        const section = document.createElement('section');
        section.className = 'mb-10';
        section.innerHTML = `
            <div class="flex items-baseline justify-between gap-4 mb-4 pb-3 border-b border-border">
                <h3 class="text-lg font-semibold tracking-tight flex items-center gap-2">
                    <i class="${iconClass} text-sm text-accent"></i> ${label}
                </h3>
                <span class="text-xs text-muted">${count}</span>
            </div>
            <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-3" id="${gridId}"></div>
        `;
        favoritesDiv.appendChild(section);
        return section.querySelector('#' + gridId);
    };

    // Favorite Venues
    if (favorites.venues && favorites.venues.length > 0) {
        const grid = renderSection('Orte', 'fas fa-map-marker-alt', favorites.venues.length, 'favorite-venues');
        favorites.venues.forEach(venueId => {
            const venue = App.data.venues.find(v => v.id === venueId);
            if (venue) grid.appendChild(createFavoriteVenueCard(venue));
        });
    }

    // Favorite Events
    if (favorites.events && favorites.events.length > 0) {
        const grid = renderSection('Veranstaltungen', 'fas fa-calendar', favorites.events.length, 'favorite-events');
        favorites.events.forEach(eventId => {
            const event = App.data.events.find(e => e.id === eventId);
            if (event) grid.appendChild(createEventCard(event));
        });
    }

    // Favorite Exhibitions
    if (favorites.exhibitions && favorites.exhibitions.length > 0) {
        const grid = renderSection('Ausstellungen', 'fas fa-palette', favorites.exhibitions.length, 'favorite-exhibitions');
        favorites.exhibitions.forEach(exhibitionId => {
            const exhibition = App.data.exhibitions.find(ex => ex.id === exhibitionId);
            if (exhibition) grid.appendChild(createFavoriteExhibitionCard(exhibition));
        });
    }
    
    console.log('✓ Favorites loaded');
}

// Create Favorite Venue Card
function createFavoriteVenueCard(venue) {
    const div = document.createElement('div');
    div.className = 'group cursor-pointer rounded-xl border border-border bg-surface hover:border-accent hover:shadow-soft p-4 transition flex flex-col';

    div.innerHTML = `
        <h4 class="text-[15px] font-semibold leading-snug group-hover:text-accent transition">${venue.name}</h4>
        <div class="mt-1 text-xs text-muted flex items-center gap-1.5">
            <i class="fas fa-location-dot text-[10px] opacity-70"></i> ${venue.address.city}
        </div>
        <div class="mt-2 flex items-center gap-3 text-xs text-muted">
            <span><i class="fas fa-calendar text-[10px] mr-1 opacity-70"></i>${venue.eventCount} Events</span>
            <span><i class="fas fa-palette text-[10px] mr-1 opacity-70"></i>${venue.exhibitionCount} Ausstellungen</span>
        </div>
        <div class="mt-3 pt-3 border-t border-border flex gap-2">
            <button class="btn-icon active w-8 h-8 inline-flex items-center justify-center rounded-md border border-border transition" onclick="toggleFavorite('${venue.id}', 'venue'); event.stopPropagation();" title="Aus Favoriten entfernen">
                <i class="fas fa-heart text-xs"></i>
            </button>
            <button class="w-8 h-8 inline-flex items-center justify-center rounded-md border border-border text-muted hover:text-accent hover:border-accent transition" onclick="centerMapOnVenue('${venue.id}'); event.stopPropagation();" title="Auf Karte zeigen">
                <i class="fas fa-map text-xs"></i>
            </button>
        </div>
    `;

    div.addEventListener('click', () => showVenueDetails(venue.id));

    return div;
}

// Create Favorite Exhibition Card
function createFavoriteExhibitionCard(exhibition) {
    const div = document.createElement('div');
    div.className = 'group cursor-pointer rounded-xl border border-border bg-surface hover:border-accent hover:shadow-soft p-4 transition flex flex-col';

    div.innerHTML = `
        <h4 class="text-[15px] font-semibold leading-snug group-hover:text-accent transition">${exhibition.title}</h4>
        ${exhibition.artist ? `<div class="mt-1 text-xs text-muted italic flex items-center gap-1.5"><i class="fas fa-user text-[10px] opacity-70"></i> ${exhibition.artist}</div>` : ''}
        <div class="mt-1 text-xs text-muted flex items-center gap-1.5">
            <i class="fas fa-location-dot text-[10px] opacity-70"></i> ${exhibition.venueName || 'Ort nicht angegeben'}
        </div>
        ${exhibition.category ? `<span class="mt-3 inline-flex w-fit items-center px-2 py-0.5 rounded-md bg-accent/10 text-accent text-[11px] font-medium">${exhibition.category}</span>` : ''}
        <div class="mt-3 pt-3 border-t border-border flex">
            <button class="btn-icon active ml-auto w-8 h-8 inline-flex items-center justify-center rounded-md border border-border transition" onclick="toggleFavorite('${exhibition.id}', 'exhibition'); event.stopPropagation();" title="Aus Favoriten entfernen">
                <i class="fas fa-heart text-xs"></i>
            </button>
        </div>
    `;

    div.addEventListener('click', () => {
        showExhibitionDetails(exhibition.id);
    });

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
            if (typeof showSuccess === 'function') {
                showSuccess('Favoriten erfolgreich importiert!');
            }
        } catch (error) {
            console.error('Error importing favorites:', error);
            if (typeof showError === 'function') {
                showError('Fehler beim Importieren der Favoriten.');
            }
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
