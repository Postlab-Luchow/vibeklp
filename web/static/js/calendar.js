// Calendar Module
async function loadCalendar() {
    console.log('Loading calendar...');
    
    try {
        const response = await fetch(`${API_BASE}/calendar`);
        const data = await response.json();
        
        const calendarDiv = document.getElementById('calendar');
        calendarDiv.innerHTML = '';
        
        if (!data.calendar || Object.keys(data.calendar).length === 0) {
            calendarDiv.innerHTML = `
                <div class="favorites-empty">
                    <i class="fas fa-calendar-times"></i>
                    <h3>Keine Veranstaltungen gefunden</h3>
                    <p>Bitte passen Sie Ihre Filter an.</p>
                </div>
            `;
            return;
        }
        
        // Sort dates
        const sortedDates = Object.keys(data.calendar).sort();
        
        sortedDates.forEach(date => {
            const day = data.calendar[date];
            const dayDiv = createCalendarDay(day);
            calendarDiv.appendChild(dayDiv);
        });
        
        console.log('✓ Calendar loaded');
    } catch (error) {
        console.error('Error loading calendar:', error);
    }
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
