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
        <div class="popup-content">
            <h3>${venue.name}</h3>
            <p>${venue.address.city}</p>
            
            ${venue.bikeRoute ? `<p><i class="fas fa-bicycle"></i> ${venue.bikeRoute}</p>` : ''}
            
            <div class="meta">
                <span><i class="fas fa-calendar"></i> ${venue.eventCount} Events</span>
                <span><i class="fas fa-palette"></i> ${venue.exhibitionCount} Ausstellungen</span>
            </div>
            
            <div class="popup-actions">
                <button class="btn-primary" onclick="showVenueDetails('${venue.id}')">
                    Details
                </button>
                <button class="btn-icon ${isFav ? 'active' : ''}" onclick="toggleFavorite('${venue.id}', 'venue'); event.stopPropagation();">
                    <i class="fas fa-heart"></i>
                </button>
            </div>
        </div>
    `;
}

// Get Venue Color
function getVenueColor(venue) {
    if (venue.eventCount > 10) return '#FF6B9D';
    if (venue.eventCount > 5) return '#FFA07A';
    return '#4ECDC4';
}

// Center Map on Venue
function centerMapOnVenue(venueId) {
    const venue = App.data.venues.find(v => v.id === venueId);
    if (!venue || !venue.coordinates) return;
    
    // Switch to map view
    switchView('map');
    
    // Close modal
    document.getElementById('detail-modal').classList.remove('active');
    
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

// Add custom marker styles
const style = document.createElement('style');
style.textContent = `
    .custom-marker {
        background: none;
        border: none;
    }
    
    .marker-pin {
        width: 30px;
        height: 40px;
        border-radius: 50% 50% 50% 0;
        background: #FF6B9D;
        position: absolute;
        transform: rotate(-45deg);
        left: 50%;
        top: 50%;
        margin: -20px 0 0 -15px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: white;
        font-size: 16px;
        box-shadow: 0 2px 8px rgba(0,0,0,0.3);
    }
    
    .marker-pin i {
        transform: rotate(45deg);
    }
    
    .marker-pin::after {
        content: '';
        width: 8px;
        height: 8px;
        margin: 20px 0 0 -4px;
        background: #fff;
        position: absolute;
        border-radius: 50%;
    }
    
    .leaflet-popup-content {
        margin: 0;
        padding: 0;
    }
    
    .custom-popup .leaflet-popup-content-wrapper {
        padding: 0;
        border-radius: 12px;
        overflow: hidden;
    }
`;
document.head.appendChild(style);

// Export functions
window.initMap = initMap;
window.updateMap = updateMap;
window.centerMapOnVenue = centerMapOnVenue;
