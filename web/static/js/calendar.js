// Calendar Module
function loadCalendar() {
    console.log('Loading calendar...');

    const calendarDiv = document.getElementById('calendar');
    calendarDiv.innerHTML = '';

    const events = filterEventsForCalendar();

    if (events.length === 0) {
        calendarDiv.innerHTML = `
            <div class="favorites-empty">
                <i class="fas fa-calendar-times"></i>
                <h3>Keine Veranstaltungen gefunden</h3>
                <p>Bitte passen Sie Ihre Filter an.</p>
            </div>
        `;
        return;
    }

    // Group filtered events by date
    const byDate = new Map();
    for (const e of events) {
        if (!e.date) continue;
        const list = byDate.get(e.date);
        if (list) list.push(e); else byDate.set(e.date, [e]);
    }

    const sortedDates = [...byDate.keys()].sort();

    sortedDates.forEach(date => {
        const dateEvents = byDate.get(date);
        const day = {
            date: date,
            dayOfWeek: getDayOfWeek(date),
            eventCount: dateEvents.length,
            events: dateEvents
        };
        const dayDiv = createCalendarDay(day);
        calendarDiv.appendChild(dayDiv);
    });

    console.log(`✓ Calendar loaded (${events.length} events)`);
}

// Apply current filter state to events for calendar rendering
function filterEventsForCalendar() {
    const { search, date, category, bikeRoute } = App.state.filters;
    const searchLower = search ? search.toLowerCase() : '';

    const venuesById = new Map(App.data.venues.map(v => [v.id, v]));

    return App.data.events.filter(event => {
        const venue = venuesById.get(event.venueId);

        // Date filter
        if (date && event.date !== date) return false;

        // Bike route filter — venue must be on a bike route
        if (bikeRoute && (!venue || !venue.bikeRoute)) return false;

        // Category filter — venue must offer this facility
        if (category) {
            const cats = (venue && venue.categories) || [];
            const match = cats.find(c => c.name === category);
            if (!match) return false;
            if (date && match.dates && match.dates.length > 0 && !match.dates.includes(date)) {
                return false;
            }
        }

        // Search filter — event fields OR venue fields
        if (searchLower) {
            const eventMatches =
                event.title.toLowerCase().includes(searchLower) ||
                (event.description && event.description.toLowerCase().includes(searchLower)) ||
                (event.category && event.category.toLowerCase().includes(searchLower));

            const venueMatches = venue && (
                venue.name.toLowerCase().includes(searchLower) ||
                (venue.description && venue.description.toLowerCase().includes(searchLower)) ||
                (venue.address && venue.address.city && venue.address.city.toLowerCase().includes(searchLower)) ||
                (venue.address && venue.address.street && venue.address.street.toLowerCase().includes(searchLower))
            );

            if (!eventMatches && !venueMatches) return false;
        }

        return true;
    });
}

function getDayOfWeek(dateStr) {
    const days = ['Sonntag', 'Montag', 'Dienstag', 'Mittwoch', 'Donnerstag', 'Freitag', 'Samstag'];
    return days[new Date(dateStr).getDay()];
}

// Create Calendar Day
function createCalendarDay(day) {
    const div = document.createElement('div');
    div.className = 'calendar-day';

    div.innerHTML = `
        <h3>
            <span>${day.dayOfWeek}, ${formatDate(day.date)}</span>
            <span class="badge">${day.eventCount} Events</span>
        </h3>
        <div class="event-grid" id="events-${day.date}"></div>
    `;

    const eventGrid = div.querySelector(`#events-${day.date}`);

    day.events.forEach(event => {
        const eventCard = createEventCard(event);
        eventGrid.appendChild(eventCard);
    });

    return div;
}

// Create Event Card
function createEventCard(event) {
    const div = document.createElement('div');
    div.className = 'event-card';

    const isFav = isFavorite(event.id);

    div.innerHTML = `
        <div class="time">${event.startTime || 'Ganztägig'}</div>
        <h4>${event.title}</h4>
        <div class="venue">
            <i class="fas fa-map-marker-alt"></i> ${event.venueName || 'Ort nicht angegeben'}
        </div>
        ${event.category ? `<span class="category">${event.category}</span>` : ''}
        ${event.admission ? `<div style="margin-top: 0.5rem; font-size: 0.85rem;"><i class="fas fa-ticket-alt"></i> ${event.admission}</div>` : ''}
        <div class="popup-actions" style="margin-top: 1rem;">
            <button class="btn-icon ${isFav ? 'active' : ''}" onclick="toggleFavorite('${event.id}', 'event'); event.stopPropagation();" title="Zu Favoriten hinzufügen">
                <i class="fas fa-heart"></i>
            </button>
        </div>
    `;

    div.addEventListener('click', () => showEventDetails(event.id));

    return div;
}

// Export functions
window.loadCalendar = loadCalendar;
