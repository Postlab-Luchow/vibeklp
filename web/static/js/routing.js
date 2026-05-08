// Routing Module
let routingControl = null;
let routingMode = false;
let selectedWaypoints = [];

// Initialize Routing
document.getElementById('route-btn')?.addEventListener('click', toggleRoutingMode);

// Toggle Routing Mode
function toggleRoutingMode() {
    routingMode = !routingMode;
    const btn = document.getElementById('route-btn');
    
    if (routingMode) {
        btn.classList.add('active');
        btn.style.background = '#FF6B9D';
        btn.style.color = 'white';
        enableRoutingMode();
    } else {
        btn.classList.remove('active');
        btn.style.background = '';
        btn.style.color = '';
        disableRoutingMode();
    }
}

// Enable Routing Mode
function enableRoutingMode() {
    console.log('Routing mode enabled');
    selectedWaypoints = [];
    
    // Show instruction
    showRoutingInstructions();
    
    // Add click handler to markers
    App.markers.eachLayer(marker => {
        marker.on('click', handleMarkerClickForRouting);
    });
}

// Disable Routing Mode
function disableRoutingMode() {
    console.log('Routing mode disabled');
    selectedWaypoints = [];
    
    // Remove routing control
    if (routingControl) {
        App.map.removeControl(routingControl);
        routingControl = null;
    }
    
    // Remove click handlers
    App.markers.eachLayer(marker => {
        marker.off('click', handleMarkerClickForRouting);
    });
    
    // Hide instructions
    hideRoutingInstructions();
}

// Handle Marker Click for Routing
function handleMarkerClickForRouting(e) {
    if (!routingMode) return;
    
    const latlng = e.latlng;
    const venue = App.data.venues.find(v => v.id === e.target.venueId);
    
    if (!venue) return;
    
    // Add waypoint
    selectedWaypoints.push({
        latlng: latlng,
        name: venue.name
    });
    
    console.log(`Added waypoint: ${venue.name}`);
    
    // Update instructions
    updateRoutingInstructions();
    
    // If we have at least 2 waypoints, calculate route
    if (selectedWaypoints.length >= 2) {
        calculateRoute();
    }
}

// Calculate Route
function calculateRoute() {
    // Remove existing routing control
    if (routingControl) {
        App.map.removeControl(routingControl);
    }
    
    // Create waypoints for Leaflet Routing Machine
    const waypoints = selectedWaypoints.map(wp => L.latLng(wp.latlng.lat, wp.latlng.lng));
    
    // Create routing control
    routingControl = L.Routing.control({
        waypoints: waypoints,
        router: L.Routing.osrmv1({
            serviceUrl: 'https://router.project-osrm.org/route/v1',
            profile: 'bike' // Use bike routing profile
        }),
        lineOptions: {
            styles: [
                { color: '#FF6B9D', opacity: 0.8, weight: 6 }
            ]
        },
        createMarker: function(i, waypoint, n) {
            const marker = L.marker(waypoint.latLng, {
                draggable: false,
                icon: L.divIcon({
                    className: 'route-marker',
                    html: `<div class="marker-pin" style="background-color: #FF6B9D;">
                               <span style="transform: rotate(45deg); display: block;">${i + 1}</span>
                           </div>`,
                    iconSize: [30, 40],
                    iconAnchor: [15, 40]
                })
            });

            marker.bindPopup(selectedWaypoints[i].name);
            return marker;
        },
        position: 'bottomleft',
        show: true,
        collapsible: true,
        routeWhileDragging: false
    }).addTo(App.map);
    
    // Listen for route found
    routingControl.on('routesfound', function(e) {
        const routes = e.routes;
        const summary = routes[0].summary;
        
        // Convert distance to km
        const distanceKm = (summary.totalDistance / 1000).toFixed(2);
        const timeMin = Math.round(summary.totalTime / 60);
        
        console.log(`Route calculated: ${distanceKm} km, ${timeMin} min`);
        
        showRouteInfo(distanceKm, timeMin);
    });
}

// Show Routing Instructions
function showRoutingInstructions() {
    const div = document.createElement('div');
    div.id = 'routing-instructions';
    div.style.cssText = `
        position: absolute;
        top: 60px;
        right: 1rem;
        background: white;
        padding: 1rem;
        border-radius: 12px;
        box-shadow: 0 2px 8px rgba(0,0,0,0.1);
        z-index: 500;
        max-width: 300px;
    `;
    
    div.innerHTML = `
        <button class="routing-close-btn" onclick="toggleRoutingMode()" title="Routenplanung schließen" aria-label="Schließen">
            <i class="fas fa-times"></i>
        </button>
        <h4 style="margin: 0 0 0.5rem 0; color: #FF6B9D; padding-right: 1.5rem;">
            <i class="fas fa-route"></i> Routenplanung
        </h4>
        <p style="margin: 0; font-size: 0.9rem;">
            Klicken Sie auf Orte auf der Karte, um eine Route zu planen.
        </p>
        <div id="waypoints-list" style="margin-top: 0.5rem; font-size: 0.85rem;"></div>
        <button onclick="clearRoute()" style="margin-top: 0.5rem; width: 100%; padding: 0.5rem; background: #f8f9fa; border: 1px solid #dee2e6; border-radius: 6px; cursor: pointer;">
            <i class="fas fa-trash"></i> Route löschen
        </button>
    `;
    
    document.querySelector('.content').appendChild(div);
}

// Hide Routing Instructions
function hideRoutingInstructions() {
    const div = document.getElementById('routing-instructions');
    if (div) {
        div.remove();
    }
}

// Update Routing Instructions
function updateRoutingInstructions() {
    const list = document.getElementById('waypoints-list');
    if (!list) return;
    
    list.innerHTML = selectedWaypoints.map((wp, i) => `
        <div style="padding: 0.25rem 0;">
            <strong>${i + 1}.</strong> ${wp.name}
        </div>
    `).join('');
}

// Show Route Info
function showRouteInfo(distanceKm, timeMin) {
    const div = document.getElementById('routing-instructions');
    if (!div) return;
    
    // Remove old info if exists
    const oldInfo = div.querySelector('.route-info');
    if (oldInfo) {
        oldInfo.remove();
    }

    const infoDiv = document.createElement('div');
    infoDiv.className = 'route-info';
    infoDiv.style.cssText = `
        margin-top: 0.5rem;
        padding: 0.5rem;
        background: #e7f5ff;
        border-radius: 6px;
        font-size: 0.85rem;
    `;

    infoDiv.innerHTML = `
        <div><i class="fas fa-bicycle"></i> <strong>${distanceKm} km</strong></div>
        <div><i class="fas fa-clock"></i> ca. ${timeMin} Minuten</div>
    `;

    div.appendChild(infoDiv);
}

// Clear Route
function clearRoute() {
    selectedWaypoints = [];

    if (routingControl) {
        App.map.removeControl(routingControl);
        routingControl = null;
    }

    const div = document.getElementById('routing-instructions');
    if (div) {
        const oldInfo = div.querySelector('.route-info');
        if (oldInfo) oldInfo.remove();
    }

    updateRoutingInstructions();

    console.log('Route cleared');
}

// Export functions
window.toggleRoutingMode = toggleRoutingMode;
window.clearRoute = clearRoute;
