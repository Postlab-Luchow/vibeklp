// Calendar Module
function loadCalendar() {
    console.log('Loading calendar...');

    const calendarDiv = document.getElementById('calendar');
    calendarDiv.innerHTML = '';

    const events = filterEventsForCalendar();

    if (events.length === 0) {
        calendarDiv.innerHTML = `
            <div class="text-center py-20 px-6 text-muted">
                <i class="fas fa-calendar-times text-4xl opacity-30 block mb-4"></i>
                <h3 class="text-base font-medium text-ink">Keine Veranstaltungen gefunden</h3>
                <p class="text-sm mt-1">Bitte passen Sie Ihre Filter an.</p>
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
    const { search, date, category, eventCategory, bikeRoute } = App.state.filters;
    const searchLower = search ? search.toLowerCase() : '';

    const venuesById = new Map(App.data.venues.map(v => [v.id, v]));

    return App.data.events.filter(event => {
        const venue = venuesById.get(event.venueId);

        // Date filter
        if (date && event.date !== date) return false;

        // Bike route filter — venue must be on a bike route
        if (bikeRoute && (!venue || !venue.bikeRoute)) return false;

        // Venue-amenity filter — venue must offer this facility
        if (category) {
            const cats = (venue && venue.categories) || [];
            const match = cats.find(c => c.name === category);
            if (!match) return false;
            if (date && match.dates && match.dates.length > 0 && !match.dates.includes(date)) {
                return false;
            }
        }

        // Event-category filter — event itself must match
        if (eventCategory && event.category !== eventCategory) return false;

        // Search filter — event fields OR venue fields
        if (searchLower) {
            const eventMatches =
                event.title.toLowerCase().includes(searchLower) ||
                (event.description && event.description.toLowerCase().includes(searchLower)) ||
                (event.category && event.category.toLowerCase().includes(searchLower)) ||
                (event.artist && event.artist.toLowerCase().includes(searchLower));

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
    div.className = 'calendar-day mb-10';

    div.innerHTML = `
        <div class="flex items-baseline justify-between gap-4 mb-4 pb-3 border-b border-border">
            <h3 class="text-lg font-semibold tracking-tight">${formatDate(day.date)}</h3>
            <span class="text-xs text-muted">${day.eventCount} Veranstaltung${day.eventCount === 1 ? '' : 'en'}</span>
        </div>
        <div class="grid sm:grid-cols-2 lg:grid-cols-3 gap-3" id="events-${day.date}"></div>
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
    div.className = 'event-card group cursor-pointer rounded-xl border border-border bg-surface hover:border-accent hover:shadow-soft p-4 transition flex flex-col';

    const isFav = isFavorite(event.id);

    div.innerHTML = `
        <div class="text-xs font-semibold text-accent tracking-wide uppercase">${event.startTime || 'Ganztägig'}</div>
        <h4 class="mt-1 text-[15px] font-medium leading-snug">${event.title} ${typeof sourceBadge === 'function' ? sourceBadge(event.source) : ''}</h4>
        <div class="mt-2 text-xs text-muted flex items-center gap-1.5">
            <i class="fas fa-location-dot text-[10px] opacity-70"></i> ${event.venueName || 'Ort nicht angegeben'}
        </div>
        ${event.category ? `<span class="mt-3 inline-flex w-fit items-center px-2 py-0.5 rounded-md bg-accent/10 text-accent text-[11px] font-medium">${event.category}</span>` : ''}
        ${event.admission ? `<div class="mt-2 text-xs text-muted"><i class="fas fa-ticket-alt text-[10px] mr-1 opacity-70"></i>${event.admission}</div>` : ''}
        <div class="mt-3 pt-3 border-t border-border flex">
            <button class="btn-icon ml-auto w-8 h-8 inline-flex items-center justify-center rounded-md border border-border text-muted hover:text-accent hover:border-accent transition ${isFav ? 'active' : ''}" onclick="toggleFavorite('${event.id}', 'event'); event.stopPropagation();" title="Zu Favoriten hinzufügen">
                <i class="fas fa-heart text-xs"></i>
            </button>
        </div>
    `;

    div.addEventListener('click', () => showEventDetails(event.id));

    return div;
}

// Export functions
window.loadCalendar = loadCalendar;
