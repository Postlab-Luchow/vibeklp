import 'dart:convert';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:klp_guide/models/venue.dart';

void main() {
  group('Venue.fromJson', () {
    late Map<String, dynamic> sampleVenueJson;

    setUpAll(() {
      final raw = File(
        'test/fixtures/venues_response_sample.json',
      ).readAsStringSync();
      final envelope = jsonDecode(raw) as Map<String, dynamic>;
      sampleVenueJson =
          (envelope['venues'] as List).first as Map<String, dynamic>;
    });

    test('parses required scalar fields', () {
      final venue = Venue.fromJson(sampleVenueJson);
      expect(venue.id, 'venue-4044317c59e8f143');
      expect(venue.name, 'SCHLANZE');
      expect(venue.description, 'Jörn (Bif) Ebersbach');
      expect(venue.bikeRoute, '5');
      expect(venue.source, 'klp');
      expect(venue.eventCount, 14);
      expect(venue.exhibitionCount, 5);
    });

    test('parses nested Address', () {
      final venue = Venue.fromJson(sampleVenueJson);
      expect(venue.address.street, 'Schlanze 2');
      expect(venue.address.postalCode, '29496');
      expect(venue.address.city, 'Waddeweitz');
    });

    test('parses nested Coordinates', () {
      final venue = Venue.fromJson(sampleVenueJson);
      expect(venue.coordinates.lat, closeTo(52.979339, 1e-6));
      expect(venue.coordinates.lng, closeTo(10.9805823, 1e-6));
    });

    test('parses empty Contact when API strips phone/email', () {
      final venue = Venue.fromJson(sampleVenueJson);
      expect(venue.contact.phone, isEmpty);
      expect(venue.contact.email, isEmpty);
      expect(venue.contact.website, isEmpty);
    });

    test('parses categories with optional dates', () {
      final venue = Venue.fromJson(sampleVenueJson);
      expect(venue.categories, hasLength(2));
      expect(venue.categories[0].name, 'Café (Kuchen, Getränke)');
      expect(venue.categories[0].dates, ['2026-05-14', '2026-05-15']);
      expect(venue.categories[1].name, 'WC');
      expect(venue.categories[1].dates, isEmpty);
    });

    test('round-trips back to the original JSON via toJson', () {
      final venue = Venue.fromJson(sampleVenueJson);
      final round = venue.toJson();
      // Required fields preserved
      expect(round['id'], 'venue-4044317c59e8f143');
      expect(round['name'], 'SCHLANZE');
      expect(round['coordinates'], {'lat': 52.979339, 'lng': 10.9805823});
    });

    test('tolerates missing optional fields', () {
      final minimal = {
        'id': 'v1',
        'name': 'Test',
        'address': {'street': '', 'postalCode': '', 'city': 'X'},
        'coordinates': {'lat': 53.0, 'lng': 11.0},
        'contact': <String, dynamic>{},
        'eventCount': 0,
        'exhibitionCount': 0,
      };
      final venue = Venue.fromJson(minimal);
      expect(venue.description, isEmpty);
      expect(venue.categories, isEmpty);
      expect(venue.bikeRoute, isEmpty);
      expect(venue.source, isEmpty);
    });
  });
}
