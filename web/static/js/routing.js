// Routing Module
let routingControl = null;
let routingMode = false;
let selectedWaypoints = [];
let lastRouteInfo = null;

// Initialize Routing
document.getElementById('route-btn')?.addEventListener('click', toggleRoutingMode);

// Toggle Routing Mode — closing minimizes (keeps the route on the map)
function toggleRoutingMode() {
    routingMode = !routingMode;
    const btn = document.getElementById('route-btn');

    if (routingMode) {
        btn.classList.add('active');
        enableRoutingMode();
    } else {
        btn.classList.remove('active');
        minimizeRoutingMode();
    }
}

// Enable Routing Mode — show the dialog and accept new waypoint clicks
function enableRoutingMode() {
    console.log('Routing mode enabled');

    showRoutingInstructions();
    updateRoutingInstructions();

    if (lastRouteInfo) {
        showRouteInfo(lastRouteInfo.distanceKm, lastRouteInfo.timeMin);
    }

    App.markers.eachLayer(marker => {
        marker.on('click', handleMarkerClickForRouting);
    });
}

// Minimize Routing Mode — hide only the dialog; keep the route + LRM box visible
function minimizeRoutingMode() {
    console.log('Routing mode minimized');

    App.markers.eachLayer(marker => {
        marker.off('click', handleMarkerClickForRouting);
    });

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
                { color: 'rgb(214, 51, 132)', opacity: 0.8, weight: 6 }
            ]
        },
        createMarker: function(i, waypoint, n) {
            const marker = L.marker(waypoint.latLng, {
                draggable: false,
                icon: L.divIcon({
                    className: 'custom-marker',
                    html: `<div class="marker-pin"><span>${i + 1}</span></div>`,
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

        lastRouteInfo = { distanceKm, timeMin };
        showRouteInfo(distanceKm, timeMin);
    });
}

// Show Routing Instructions
function showRoutingInstructions() {
    const div = document.createElement('div');
    div.id = 'routing-instructions';
    div.className = 'absolute top-16 right-4 z-[500] w-[260px] max-w-[80vw] p-4 rounded-2xl bg-surface text-ink border border-border shadow-soft';

    div.innerHTML = `
        <button class="absolute top-2 right-2 w-7 h-7 inline-flex items-center justify-center rounded-full bg-surface-elevated text-muted hover:bg-accent hover:text-white text-xs transition" onclick="toggleRoutingMode()" title="Routenplanung schließen" aria-label="Schließen">
            <i class="fas fa-times"></i>
        </button>
        <h4 class="text-accent font-semibold text-sm flex items-center gap-1.5 pr-6">
            <i class="fas fa-route"></i> Routenplanung
        </h4>
        <p class="mt-1 text-xs text-muted leading-relaxed">
            Klicken Sie auf Orte auf der Karte, um eine Route zu planen.
        </p>
        <div id="waypoints-list" class="mt-2 text-xs space-y-0.5"></div>
        <button onclick="clearRoute()" class="mt-3 w-full px-3 py-1.5 rounded-md bg-surface-elevated hover:bg-border text-xs transition">
            <i class="fas fa-trash mr-1"></i> Route löschen
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
        <div class="py-0.5"><span class="text-muted">${i + 1}.</span> ${wp.name}</div>
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
    infoDiv.className = 'route-info mt-3 p-2.5 rounded-md bg-accent/10 text-accent text-xs space-y-0.5';
    infoDiv.innerHTML = `
        <div><i class="fas fa-bicycle"></i> <strong>${distanceKm} km</strong></div>
        <div><i class="fas fa-clock"></i> ca. ${timeMin} Minuten</div>
    `;

    div.appendChild(infoDiv);
}

// Clear Route
function clearRoute() {
    selectedWaypoints = [];
    lastRouteInfo = null;

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
