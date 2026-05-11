// Map Module
function initMap() {
    console.log('Initializing map...');
    
    // Create map centered on Wendland region
    App.map = L.map('map').setView([53.0, 11.2], 10);
    
    // Add OpenStreetMap tiles
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
        attribution: '© <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
        maxZoom: 18
    }).addTo(App.map);
    
    // Create marker cluster group
    App.markers = L.markerClusterGroup({
        maxClusterRadius: 50,
        spiderfyOnMaxZoom: true,
        showCoverageOnHover: false,
        zoomToBoundsOnClick: true
    });
    
    // Add markers to map
    updateMap();
    
    // Add locate control
    document.getElementById('locate-btn').addEventListener('click', locateUser);
    
    console.log('✓ Map initialized');
}

// Update Map with Filtered Venues
function updateMap() {
    if (!App.map || !App.markers) return;
    
    // Clear existing markers
    App.markers.clearLayers();
    
    // Add markers for filtered venues
    App.data.filteredVenues.forEach(venue => {
        if (venue.coordinates && venue.coordinates.lat && venue.coordinates.lng) {
            const marker = createVenueMarker(venue);
            App.markers.addLayer(marker);
        }
    });
    
    // Add marker cluster to map
    App.map.addLayer(App.markers);
    
    // Fit bounds if there are markers
    if (App.data.filteredVenues.length > 0) {
        const bounds = App.markers.getBounds();
        if (bounds.isValid()) {
            App.map.fitBounds(bounds, { padding: [50, 50] });
        }
    }
}

// Create Venue Marker
function createVenueMarker(venue) {
    // Custom icon based on venue type
    const icon = L.divIcon({
        className: 'custom-marker',
        html: `<div class="marker-pin" style="background-color: ${getVenueColor(venue)};">
                   <i class="fas fa-map-marker-alt"></i>
               </div>`,
        iconSize: [30, 40],
        iconAnchor: [15, 40],
        popupAnchor: [0, -40]
    });
    
    const marker = L.marker([venue.coordinates.lat, venue.coordinates.lng], { icon });
    
    // Create popup content
    const popupContent = createPopupContent(venue);
    marker.bindPopup(popupContent, {
        maxWidth: 300,
        className: 'custom-popup'
    });
    
    // Store venue ID on marker
    marker.venueId = venue.id;
    
    return marker;
}

// Create Popup Content
function createPopupContent(venue) {
    const isFav = isFavorite(venue.id);

    return `
        <div class="p-4 min-w-[220px] max-w-[280px]">
            <h3 class="text-[15px] font-semibold tracking-tight text-ink leading-snug">${venue.name}</h3>
            <p class="mt-1 text-xs text-muted flex items-center gap-1.5">
                <i class="fas fa-location-dot text-[10px] opacity-70"></i> ${venue.address.city}
            </p>

            ${venue.bikeRoute ? `<p class="mt-2 text-xs text-muted"><i class="fas fa-bicycle text-[10px] mr-1 opacity-70"></i>${venue.bikeRoute}</p>` : ''}

            <div class="mt-2 flex items-center gap-3 text-xs text-muted">
                <span><i class="fas fa-calendar text-[10px] mr-1 opacity-70"></i>${venue.eventCount} Events</span>
                <span><i class="fas fa-palette text-[10px] mr-1 opacity-70"></i>${venue.exhibitionCount} Ausst.</span>
            </div>

            <div class="mt-3 flex items-center gap-2">
                <button class="inline-flex items-center gap-1 px-3 h-8 rounded-md bg-accent text-white text-xs font-medium hover:bg-accent-strong transition" onclick="showVenueDetails('${venue.id}')">
                    Details
                </button>
                <button class="btn-icon w-8 h-8 inline-flex items-center justify-center rounded-md border border-border text-muted hover:text-accent hover:border-accent transition ${isFav ? 'active' : ''}" onclick="toggleFavorite('${venue.id}', 'venue'); event.stopPropagation();" aria-label="Favorit">
                    <i class="fas fa-heart text-xs"></i>
                </button>
            </div>
        </div>
    `;
}

// Get Venue Color — pin tint based on activity. Same values for light/dark
// since the marker is over the map and uses its own bg layer.
function getVenueColor(venue) {
    if (venue.eventCount > 10) return '#D63384';
    if (venue.eventCount > 5) return '#F472B6';
    return '#5EAAA8';
}

// Center Map on Venue
function centerMapOnVenue(venueId) {
    const venue = App.data.venues.find(v => v.id === venueId);
    if (!venue || !venue.coordinates) return;
    
    // Switch to map view
    switchView('map');

    // Force-close modal (clears the modal back-stack)
    hideModal();
    
    // Center and zoom to venue
    App.map.setView([venue.coordinates.lat, venue.coordinates.lng], 15);
    
    // Find and open marker popup
    App.markers.eachLayer(marker => {
        if (marker.venueId === venueId) {
            marker.openPopup();
        }
    });
}

// Locate User
function locateUser() {
    if (!navigator.geolocation) {
        if (typeof showError === 'function') {
            showError('Geolocation wird von Ihrem Browser nicht unterstützt.');
        }
        return;
    }
    
    navigator.geolocation.getCurrentPosition(
        position => {
            const { latitude, longitude } = position.coords;
            App.map.setView([latitude, longitude], 13);
            
            // Add user marker
            L.marker([latitude, longitude], {
                icon: L.divIcon({
                    className: 'user-marker',
                    html: '<div class="marker-pin" style="background-color: #007bff;"><i class="fas fa-user"></i></div>',
                    iconSize: [30, 40],
                    iconAnchor: [15, 40]
                })
            }).addTo(App.map).bindPopup('Ihr Standort').openPopup();
        },
        error => {
            console.error('Geolocation error:', error);
            if (typeof showError === 'function') {
                showError('Standort konnte nicht ermittelt werden.');
            }
        }
    );
}

// Marker pin + popup styling lives in web/static/css/tailwind-input.css (the
// generated tailwind.css picks them up via @layer components).

// Export functions
window.initMap = initMap;
window.updateMap = updateMap;
window.centerMapOnVenue = centerMapOnVenue;
