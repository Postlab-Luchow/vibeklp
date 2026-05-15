// Global App State
const App = {
    data: {
        venues: [],
        events: [],
        exhibitions: [],
        categories: [],
        eventCategories: [],
        filteredVenues: [],
        filteredEvents: []
    },
    state: {
        currentView: 'map',
        selectedVenue: null,
        selectedEvent: null,
        modalStack: [],
        filters: {
            search: '',
            date: '',
            timeOfDay: '',
            category: '',
            eventCategory: '',
            bikeRoute: false
        }
    },
    map: null,
    markers: null
};

// API Base URL
const API_BASE = '/api';

// Source code → human label. The main KLP source is the implicit default,
// so we hide its badge to keep the UI uncluttered.
const SOURCE_LABELS = {
    wendlandpartie: 'WendlandPartie',
    landgang: 'Landgang'
};

function sourceBadge(source) {
    const label = SOURCE_LABELS[source];
    if (!label) return '';
    return `<span class="source-badge source-${source}">${label}</span>`;
}

// Section heading shared by all modal sections — small caps, muted, with
// optional icon. Kept inline so the modal renderers stay readable.
function sectionHeading(label) {
    return `<h3 class="mt-7 text-[11px] uppercase tracking-[0.12em] font-semibold text-muted">${label}</h3>`;
}

// Opens a collapsible <details> section. Caller closes with `</details>`.
// The caret rotates 180° when open via the .details-caret CSS rule.
function collapsibleSectionOpen(label) {
    return `
        <details class="mt-7" open>
            <summary class="details-summary flex items-center justify-between gap-3 cursor-pointer select-none">
                <h3 class="text-[11px] uppercase tracking-[0.12em] font-semibold text-muted">${label}</h3>
                <i class="fas fa-chevron-down text-[10px] text-muted details-caret"></i>
            </summary>
    `;
}

// Common card styling for events/exhibitions rendered inside a modal.
const MODAL_CARD_CLASS = 'cursor-pointer rounded-xl border border-border bg-surface hover:border-accent hover:shadow-soft p-4 transition';

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
        initMobileSidebar();

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
        
        // Load venue-amenity categories with retry
        try {
            const categoriesResponse = await fetchWithRetry(`${API_BASE}/categories`);
            const categoriesData = await categoriesResponse.json();
            App.data.categories = categoriesData.categories || [];
            console.log(`✓ Loaded ${App.data.categories.length} venue categories`);
        } catch (error) {
            console.error('Failed to load categories:', error);
            failedEndpoints.push('Kategorien');
        }

        // Load event categories with retry
        try {
            const ecResponse = await fetchWithRetry(`${API_BASE}/event-categories`);
            const ecData = await ecResponse.json();
            App.data.eventCategories = ecData.categories || [];
            console.log(`✓ Loaded ${App.data.eventCategories.length} event categories`);
        } catch (error) {
            console.error('Failed to load event categories:', error);
            // Non-fatal: filter will just be empty
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
    
    // Populate venue-amenity filter
    const categoryFilter = document.getElementById('category-filter');
    App.data.categories.forEach(cat => {
        const option = document.createElement('option');
        option.value = cat.name;
        option.textContent = `${cat.name} (${cat.count})`;
        categoryFilter.appendChild(option);
    });

    // Populate event-category filter — only show categories that actually have items
    const eventCategoryFilter = document.getElementById('event-category-filter');
    if (eventCategoryFilter) {
        App.data.eventCategories.forEach(cat => {
            if (!cat.count) return;
            const option = document.createElement('option');
            option.value = cat.name;
            option.textContent = `${cat.name} (${cat.count})`;
            eventCategoryFilter.appendChild(option);
        });
    }
}

// Mobile bottom-sheet wiring (no-op on desktop where the elements are display:none)
function initMobileSidebar() {
    const toggle = document.getElementById('mobile-filter-toggle');
    const closeBtn = document.getElementById('sidebar-close');
    const backdrop = document.getElementById('sidebar-backdrop');

    toggle?.addEventListener('click', openMobileSidebar);
    closeBtn?.addEventListener('click', closeMobileSidebar);
    backdrop?.addEventListener('click', closeMobileSidebar);
}

function openMobileSidebar() {
    document.body.classList.add('sidebar-open');
}

function closeMobileSidebar() {
    document.body.classList.remove('sidebar-open');
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
    closeMobileSidebar();

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

// Count events/exhibitions at a venue that pass the active filters. When no
// content filter is active we return the venue's stored totals untouched so
// the result-list and map popups stay consistent with the venue modal.
function venueFilteredCounts(venue) {
    const f = App.state.filters;
    const searchLower = (f.search || '').toLowerCase();
    const hasFilter = !!(f.date || f.timeOfDay || f.eventCategory || searchLower);

    if (!hasFilter) {
        return { eventCount: venue.eventCount, exhibitionCount: venue.exhibitionCount, isFiltered: false };
    }

    const eventCount = App.data.events.reduce((n, e) => {
        if (e.venueId !== venue.id) return n;
        if (f.date && e.date !== f.date) return n;
        if (!eventInTimeBucket(e, f.timeOfDay)) return n;
        if (f.eventCategory && e.category !== f.eventCategory) return n;
        if (searchLower) {
            const hay = e.title.toLowerCase()
                + ' ' + (e.description ? e.description.toLowerCase() : '')
                + ' ' + (e.category ? e.category.toLowerCase() : '')
                + ' ' + (e.artist ? e.artist.toLowerCase() : '');
            if (!hay.includes(searchLower)) return n;
        }
        return n + 1;
    }, 0);

    const exhibitionCount = f.timeOfDay ? 0 : App.data.exhibitions.reduce((n, ex) => {
        if (ex.venueId !== venue.id) return n;
        if (f.eventCategory && ex.category !== f.eventCategory) return n;
        if (searchLower) {
            const hay = ex.title.toLowerCase()
                + ' ' + (ex.description ? ex.description.toLowerCase() : '')
                + ' ' + (ex.artist ? ex.artist.toLowerCase() : '')
                + ' ' + (ex.category ? ex.category.toLowerCase() : '');
            if (!hay.includes(searchLower)) return n;
        }
        return n + 1;
    }, 0);

    return { eventCount, exhibitionCount, isFiltered: true };
}

// Time-of-day buckets: event qualifies if its startTime (HH:MM) falls in the bucket.
// Events without a startTime are excluded when any bucket is active.
function eventInTimeBucket(event, bucket) {
    if (!bucket) return true;
    if (!event.startTime) return false;
    const t = event.startTime;
    switch (bucket) {
        case 'morning':   return t < '13:00';
        case 'afternoon': return t >= '13:00' && t < '17:00';
        case 'evening':   return t >= '17:00' && t < '21:00';
        case 'late':      return t >= '21:00';
        default:          return true;
    }
}

// Apply Filters
function applyFilters() {
    const { search, date, timeOfDay, category, eventCategory, bikeRoute } = App.state.filters;
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
        
        // Date filter - venue must have an event on this date (and in the
        // selected time bucket, if any).
        if (date) {
            const hasEventOnDate = venueEvents.some(e =>
                e.date === date && eventInTimeBucket(e, timeOfDay)
            );
            if (!hasEventOnDate) return false;
        } else if (timeOfDay) {
            // Time-only filter — venue must have at least one event in this bucket.
            const hasEventInTime = venueEvents.some(e => eventInTimeBucket(e, timeOfDay));
            if (!hasEventInTime) return false;
        }

        // Event-category filter — venue must have at least one event or
        // exhibition in this category (and on the selected date / in the
        // selected time bucket, if any). Exhibitions have no time, so they
        // never satisfy a time filter.
        if (eventCategory) {
            const eventMatch = venueEvents.some(e =>
                e.category === eventCategory &&
                (!date || e.date === date) &&
                eventInTimeBucket(e, timeOfDay)
            );
            const exhibitionMatch = !timeOfDay && venueExhibitions.some(ex => ex.category === eventCategory);
            if (!eventMatch && !exhibitionMatch) return false;
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
                (e.category && e.category.toLowerCase().includes(searchLower)) ||
                (e.artist && e.artist.toLowerCase().includes(searchLower))
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

    // Refresh open venue modal so its events list reflects the new filter.
    // Event/exhibition modals don't depend on filters, so they're skipped.
    const topModal = App.state.modalStack[App.state.modalStack.length - 1];
    if (topModal && topModal.type === 'venue') {
        _renderVenueModal(topModal.id);
    }
}

// Modal payload cache — keyed by "type:id". Data is static for the session,
// so re-renders (e.g. after a filter change) skip the network round-trip and
// the loading spinner.
const _modalDataCache = new Map();

async function _fetchModalData(type, id, url) {
    const key = `${type}:${id}`;
    if (_modalDataCache.has(key)) return _modalDataCache.get(key);
    showLoading();
    try {
        const response = await fetch(url);
        const data = await response.json();
        _modalDataCache.set(key, data);
        return data;
    } finally {
        hideLoading();
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
            <div class="px-6 lg:px-8 py-12 text-center text-muted">
                <i class="fas fa-search text-2xl opacity-40 block mb-3"></i>
                <p class="text-sm">Keine Ergebnisse gefunden</p>
                <p class="text-xs opacity-70 mt-1">Andere Suchbegriffe oder Filter probieren</p>
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
    div.className = 'result-item group cursor-pointer px-6 lg:px-8 py-4 hover:bg-surface-elevated transition';

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
            (e.description && e.description.toLowerCase().includes(searchLower)) ||
            (e.artist && e.artist.toLowerCase().includes(searchLower))
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
            matchInfo = `<span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-md bg-accent/15 text-accent text-[11px] font-medium">
                <i class="fas fa-search text-[9px]"></i> ${matches.join(', ')}
            </span>`;
        }
    }

    const { eventCount, exhibitionCount, isFiltered } = venueFilteredCounts(venue);
    const counts = [];
    if (eventCount) counts.push(`<span><i class="fas fa-calendar text-[10px] mr-1 opacity-70"></i>${eventCount} Events</span>`);
    if (exhibitionCount) counts.push(`<span><i class="fas fa-palette text-[10px] mr-1 opacity-70"></i>${exhibitionCount} Ausstellungen</span>`);
    if (isFiltered) counts.push('<span class="text-accent" title="Gefilterte Anzahl"><i class="fas fa-filter text-[10px]"></i></span>');

    div.innerHTML = `
        <h4 class="text-[15px] font-semibold text-ink leading-snug group-hover:text-accent transition">
            ${venue.name} ${sourceBadge(venue.source)}
        </h4>
        <p class="mt-0.5 text-[13px] text-muted flex items-center gap-1.5">
            <i class="fas fa-location-dot text-[11px] opacity-70"></i>
            ${venue.address.city || ''}
        </p>
        ${counts.length || matchInfo ? `
            <div class="mt-2 flex items-center flex-wrap gap-x-3 gap-y-1 text-xs text-muted">
                ${counts.join('')}
                ${matchInfo}
            </div>
        ` : ''}
    `;

    div.addEventListener('click', () => showVenueDetails(venue.id));
    return div;
}

// Modal stack — supports back-navigation between modals (e.g. venue → event → back to venue)
async function _renderVenueModal(venueId) {
    try {
        const venue = await _fetchModalData('venue', venueId, `${API_BASE}/venues/${venueId}`);

        const modal = document.getElementById('detail-modal');
        const content = document.getElementById('detail-content');

        // Respect the active date + time + search + event-category filters so
        // the modal only shows matching events.
        const dateFilter = App.state.filters.date;
        const timeFilter = App.state.filters.timeOfDay;
        const searchFilter = (App.state.filters.search || '').toLowerCase();
        const eventCategoryFilter = App.state.filters.eventCategory;
        const allVenueEvents = venue.events || [];
        const venueEvents = allVenueEvents.filter(e => {
            if (dateFilter && e.date !== dateFilter) return false;
            if (!eventInTimeBucket(e, timeFilter)) return false;
            if (eventCategoryFilter && e.category !== eventCategoryFilter) return false;
            if (searchFilter) {
                const hay =
                    e.title.toLowerCase() +
                    ' ' + (e.description ? e.description.toLowerCase() : '') +
                    ' ' + (e.category ? e.category.toLowerCase() : '') +
                    ' ' + (e.artist ? e.artist.toLowerCase() : '');
                if (!hay.includes(searchFilter)) return false;
            }
            return true;
        }).sort((a, b) => {
            const dateCmp = (a.date || '').localeCompare(b.date || '');
            if (dateCmp !== 0) return dateCmp;
            return (a.startTime || '').localeCompare(b.startTime || '');
        });
        const isEventsFiltered = (dateFilter || timeFilter || searchFilter || eventCategoryFilter) && venueEvents.length !== allVenueEvents.length;
        const eventsHeading = dateFilter
            ? `Veranstaltungen am ${formatDate(dateFilter)} (${venueEvents.length})`
            : `Veranstaltungen (${venueEvents.length})`;

        // Group events by date so multi-day venues mirror the calendar layout.
        const venueEventsByDate = new Map();
        for (const e of venueEvents) {
            const key = e.date || '';
            const list = venueEventsByDate.get(key);
            if (list) list.push(e); else venueEventsByDate.set(key, [e]);
        }
        const venueEventDates = [...venueEventsByDate.keys()];
        const eventsGroupedByDay = venueEventDates.length > 1;

        const venueEventCard = (e, withDate) => {
            const dateline = withDate
                ? `${formatDate(e.date)}${e.startTime ? ' · ' + e.startTime : ''}`
                : (e.startTime || 'Ganztägig');
            return `
                <div class="${MODAL_CARD_CLASS}" onclick="pushModal('event', '${e.id}')">
                    <div class="text-xs font-medium text-accent">${dateline}</div>
                    <h4 class="mt-1 font-medium leading-snug">${e.title}</h4>
                    ${e.category ? `<span class="mt-2 inline-flex items-center px-2 py-0.5 rounded-md bg-accent/10 text-accent text-[11px] font-medium">${e.category}</span>` : ''}
                </div>
            `;
        };
        const eventsFilterHint = isEventsFiltered
            ? `<p class="filter-hint"><i class="fas fa-filter"></i> ${venueEvents.length} von ${allVenueEvents.length} Veranstaltungen entsprechen den aktiven Filtern.</p>`
            : '';

        // Exhibitions have no time/date — they're hidden entirely when a time
        // bucket is active; otherwise search and event-category filters apply.
        const allVenueExhibitions = venue.exhibitions || [];
        const venueExhibitions = timeFilter ? [] : allVenueExhibitions.filter(ex => {
            if (eventCategoryFilter && ex.category !== eventCategoryFilter) return false;
            if (!searchFilter) return true;
            const hay =
                ex.title.toLowerCase() +
                ' ' + (ex.description ? ex.description.toLowerCase() : '') +
                ' ' + (ex.artist ? ex.artist.toLowerCase() : '') +
                ' ' + (ex.category ? ex.category.toLowerCase() : '');
            return hay.includes(searchFilter);
        });
        const isExhibitionsFiltered = (timeFilter || searchFilter || eventCategoryFilter) && venueExhibitions.length !== allVenueExhibitions.length;
        const filterHintCls = 'mt-2 inline-flex items-center gap-2 px-3 py-1.5 rounded-md bg-warn-bg text-warn-text text-xs';
        const exhibitionsFilterHint = isExhibitionsFiltered
            ? `<p class="${filterHintCls}"><i class="fas fa-filter"></i> ${venueExhibitions.length} von ${allVenueExhibitions.length} Ausstellungen entsprechen der Suche.</p>`
            : '';

        content.innerHTML = `
            <h2 class="text-xl sm:text-2xl font-semibold tracking-tight pr-10">${venue.name} ${sourceBadge(venue.source)}</h2>
            ${venue.description ? `<p class="mt-3 text-sm text-muted leading-relaxed">${venue.description}</p>` : ''}

            ${sectionHeading('Adresse')}
            <p class="mt-2 text-sm">${venue.address.street}<br>${venue.address.postalCode} ${venue.address.city}</p>

            ${(venue.contact.phone || venue.contact.email || venue.contact.website) ? `
                <div class="mt-3 space-y-1 text-sm text-muted">
                    ${venue.contact.phone ? `<p><i class="fas fa-phone w-4 text-center text-[11px] opacity-70"></i> ${venue.contact.phone}</p>` : ''}
                    ${venue.contact.email ? `<p><i class="fas fa-envelope w-4 text-center text-[11px] opacity-70"></i> ${venue.contact.email}</p>` : ''}
                    ${venue.contact.website ? `<p><i class="fas fa-globe w-4 text-center text-[11px] opacity-70"></i> <a href="${venue.contact.website}" target="_blank" class="text-accent hover:underline">${venue.contact.website}</a></p>` : ''}
                </div>
            ` : ''}

            ${venueEvents.length > 0 ? `
                ${collapsibleSectionOpen(eventsHeading)}
                    ${eventsFilterHint ? `<p class="${filterHintCls}"><i class="fas fa-filter"></i> ${venueEvents.length} von ${allVenueEvents.length} Veranstaltungen entsprechen den aktiven Filtern.</p>` : ''}
                    ${eventsGroupedByDay
                        ? venueEventDates.map(date => {
                            const dayEvents = venueEventsByDate.get(date);
                            const countLabel = `${dayEvents.length} Veranstaltung${dayEvents.length === 1 ? '' : 'en'}`;
                            return `
                                <details class="mt-4" open>
                                    <summary class="details-summary flex items-center justify-between gap-3 mb-3 pb-2 border-b border-border cursor-pointer select-none">
                                        <h4 class="text-sm font-semibold tracking-tight">${formatDate(date)}</h4>
                                        <div class="flex items-center gap-3 shrink-0">
                                            <span class="text-xs text-muted">${countLabel}</span>
                                            <i class="fas fa-chevron-down text-[10px] text-muted details-caret"></i>
                                        </div>
                                    </summary>
                                    <div class="grid sm:grid-cols-2 gap-3">
                                        ${dayEvents.map(e => venueEventCard(e, false)).join('')}
                                    </div>
                                </details>
                            `;
                          }).join('')
                        : `<div class="mt-3 grid sm:grid-cols-2 gap-3">
                                ${venueEvents.map(e => venueEventCard(e, true)).join('')}
                           </div>`
                    }
                </details>
            ` : (allVenueEvents.length > 0 ? `
                ${sectionHeading('Veranstaltungen (0)')}
                <p class="${filterHintCls}"><i class="fas fa-filter"></i> Keine der ${allVenueEvents.length} Veranstaltungen entspricht den aktiven Filtern.</p>
            ` : '')}

            ${venueExhibitions.length > 0 ? `
                ${collapsibleSectionOpen(`Ausstellungen (${venueExhibitions.length})`)}
                    ${exhibitionsFilterHint}
                    <div class="mt-3 grid sm:grid-cols-2 gap-3">
                        ${venueExhibitions.map(ex => `
                            <div class="rounded-xl border border-border bg-surface p-4">
                                <h4 class="font-medium leading-snug">${ex.title}</h4>
                                ${ex.artist ? `<p class="mt-1 text-xs text-muted italic">${ex.artist}</p>` : ''}
                                ${ex.category ? `<span class="mt-2 inline-flex items-center px-2 py-0.5 rounded-md bg-accent/10 text-accent text-[11px] font-medium">${ex.category}</span>` : ''}
                            </div>
                        `).join('')}
                    </div>
                </details>
            ` : (allVenueExhibitions.length > 0 ? `
                ${sectionHeading('Ausstellungen (0)')}
                <p class="${filterHintCls}"><i class="fas fa-filter"></i> Keine der ${allVenueExhibitions.length} Ausstellungen entspricht der Suche.</p>
            ` : '')}

            <div class="mt-8 flex items-center gap-2">
                <button class="inline-flex items-center gap-2 px-4 h-10 rounded-lg bg-accent text-white text-sm font-medium hover:bg-accent-strong transition" onclick="centerMapOnVenue('${venue.id}')">
                    <i class="fas fa-map"></i> Auf Karte zeigen
                </button>
                <button class="btn-icon w-10 h-10 inline-flex items-center justify-center rounded-lg border border-border hover:border-accent hover:text-accent transition ${isFavorite(venue.id) ? 'active' : ''}" onclick="toggleFavorite('${venue.id}', 'venue')" aria-label="Favorit">
                    <i class="fas fa-heart"></i>
                </button>
            </div>
        `;

        modal.classList.add('active');
    } catch (error) {
        console.error('Error loading venue details:', error);
    }
}

async function _renderEventModal(eventId) {
    try {
        const event = await _fetchModalData('event', eventId, `${API_BASE}/events/${eventId}`);

        const modal = document.getElementById('detail-modal');
        const content = document.getElementById('detail-content');

        content.innerHTML = `
            <h2 class="text-xl sm:text-2xl font-semibold tracking-tight pr-10">${event.title} ${sourceBadge(event.source)}</h2>
            ${event.description ? `<p class="mt-3 text-sm text-muted leading-relaxed">${event.description}</p>` : ''}

            ${sectionHeading('Wann')}
            <p class="mt-2 text-sm">${formatDate(event.date)} ${event.startTime ? `<span class="text-muted">um ${event.startTime} Uhr</span>` : ''}</p>

            ${event.venue ? `
                ${sectionHeading('Wo')}
                <p class="mt-2 text-sm">${event.venue.name}<br>${event.venue.address.street}<br>${event.venue.address.postalCode} ${event.venue.address.city}</p>
            ` : ''}

            ${(event.artist || event.admission || event.category) ? `
                <div class="mt-4 space-y-1 text-sm text-muted">
                    ${event.artist ? `<p><i class="fas fa-user w-4 text-center text-[11px] opacity-70"></i> ${event.artist}</p>` : ''}
                    ${event.admission ? `<p><i class="fas fa-ticket-alt w-4 text-center text-[11px] opacity-70"></i> Eintritt: ${event.admission}</p>` : ''}
                    ${event.category ? `<p><i class="fas fa-tag w-4 text-center text-[11px] opacity-70"></i> ${event.category}</p>` : ''}
                </div>
            ` : ''}

            <div class="mt-8 flex items-center gap-2">
                ${event.venue ? `
                    <button class="inline-flex items-center gap-2 px-4 h-10 rounded-lg bg-accent text-white text-sm font-medium hover:bg-accent-strong transition" onclick="centerMapOnVenue('${event.venueId}')">
                        <i class="fas fa-map"></i> Auf Karte zeigen
                    </button>
                ` : ''}
                <button class="btn-icon w-10 h-10 inline-flex items-center justify-center rounded-lg border border-border hover:border-accent hover:text-accent transition ${isFavorite(eventId) ? 'active' : ''}" onclick="toggleFavorite('${eventId}', 'event')" aria-label="Favorit">
                    <i class="fas fa-heart"></i>
                </button>
            </div>
        `;

        modal.classList.add('active');
    } catch (error) {
        console.error('Error loading event details:', error);
    }
}

async function _renderExhibitionModal(exhibitionId) {
    try {
        const exhibition = await _fetchModalData('exhibition', exhibitionId, `${API_BASE}/exhibitions/${exhibitionId}`);

        const modal = document.getElementById('detail-modal');
        const content = document.getElementById('detail-content');

        content.innerHTML = `
            <h2 class="text-xl sm:text-2xl font-semibold tracking-tight pr-10">${exhibition.title}</h2>
            ${exhibition.artist ? `<p class="mt-2 text-sm text-muted italic"><i class="fas fa-palette text-[11px] mr-1 opacity-70"></i> ${exhibition.artist}</p>` : ''}
            ${exhibition.description ? `<p class="mt-3 text-sm text-muted leading-relaxed">${exhibition.description}</p>` : ''}

            ${exhibition.venue ? `
                ${sectionHeading('Wo')}
                <p class="mt-2 text-sm">${exhibition.venue.name}<br>${exhibition.venue.address.street}<br>${exhibition.venue.address.postalCode} ${exhibition.venue.address.city}</p>
            ` : ''}

            ${exhibition.category ? `<p class="mt-4 text-sm text-muted"><i class="fas fa-tag w-4 text-center text-[11px] opacity-70"></i> ${exhibition.category}</p>` : ''}

            <div class="mt-8 flex items-center gap-2">
                ${exhibition.venue ? `
                    <button class="inline-flex items-center gap-2 px-4 h-10 rounded-lg bg-accent text-white text-sm font-medium hover:bg-accent-strong transition" onclick="centerMapOnVenue('${exhibition.venueId}')">
                        <i class="fas fa-map"></i> Auf Karte zeigen
                    </button>
                ` : ''}
                <button class="btn-icon w-10 h-10 inline-flex items-center justify-center rounded-lg border border-border hover:border-accent hover:text-accent transition ${isFavorite(exhibitionId) ? 'active' : ''}" onclick="toggleFavorite('${exhibitionId}', 'exhibition')" aria-label="Favorit">
                    <i class="fas fa-heart"></i>
                </button>
            </div>
        `;

        modal.classList.add('active');
    } catch (error) {
        console.error('Error loading exhibition details:', error);
    }
}

function _renderTopOfModalStack() {
    const top = App.state.modalStack[App.state.modalStack.length - 1];
    if (!top) return;
    if (top.type === 'venue') return _renderVenueModal(top.id);
    if (top.type === 'event') return _renderEventModal(top.id);
    if (top.type === 'exhibition') return _renderExhibitionModal(top.id);
}

// Public entry points (always reset the modal stack)
function showVenueDetails(venueId) {
    closeMobileSidebar();
    App.state.modalStack = [{ type: 'venue', id: venueId }];
    return _renderVenueModal(venueId);
}

function showEventDetails(eventId) {
    closeMobileSidebar();
    App.state.modalStack = [{ type: 'event', id: eventId }];
    return _renderEventModal(eventId);
}

function showExhibitionDetails(exhibitionId) {
    closeMobileSidebar();
    App.state.modalStack = [{ type: 'exhibition', id: exhibitionId }];
    return _renderExhibitionModal(exhibitionId);
}

// Push a new modal on top of the stack — used when navigating from one modal to another
function pushModal(type, id) {
    App.state.modalStack.push({ type, id });
    return _renderTopOfModalStack();
}

// Close the topmost modal — pops back to the previous one if any, otherwise hides
function closeModal() {
    App.state.modalStack.pop();
    if (App.state.modalStack.length > 0) {
        return _renderTopOfModalStack();
    }
    hideModal();
}

// Force-hide the modal and clear navigation history (used by "Auf Karte zeigen" etc.)
function hideModal() {
    App.state.modalStack = [];
    document.getElementById('detail-modal').classList.remove('active');
}

document.querySelector('.modal-close')?.addEventListener('click', closeModal);

document.getElementById('detail-modal')?.addEventListener('click', (e) => {
    if (e.target.id === 'detail-modal') closeModal();
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
    errorDiv.className = 'alert-enter flex items-center gap-3 px-4 py-3 rounded-xl bg-danger-bg text-danger-text shadow-soft border border-danger-text/20';
    errorDiv.innerHTML = `
        <i class="fas fa-exclamation-circle text-base shrink-0"></i>
        <span class="text-sm flex-1">${escapeHtml(message)}</span>
        ${options.retry ? `<button class="text-xs font-medium px-3 h-8 rounded-md bg-danger-text text-white hover:opacity-90 transition" onclick="${options.retry}">Erneut versuchen</button>` : ''}
        <button class="w-7 h-7 inline-flex items-center justify-center rounded-md hover:bg-danger-text/10 transition" onclick="dismissError('${errorId}')" aria-label="Fehler schließen">
            <i class="fas fa-times text-xs"></i>
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
    successDiv.className = 'alert-enter flex items-center gap-3 px-4 py-3 rounded-xl bg-success-bg text-success-text shadow-soft border border-success-text/20';
    successDiv.innerHTML = `
        <i class="fas fa-check-circle text-base shrink-0"></i>
        <span class="text-sm flex-1">${escapeHtml(message)}</span>
        <button class="w-7 h-7 inline-flex items-center justify-center rounded-md hover:bg-success-text/10 transition" onclick="dismissError('${successId}')" aria-label="Nachricht schließen">
            <i class="fas fa-times text-xs"></i>
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
window.pushModal = pushModal;
window.closeModal = closeModal;
window.hideModal = hideModal;
window.showError = showError;
window.showSuccess = showSuccess;
window.dismissError = dismissError;
