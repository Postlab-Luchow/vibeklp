import 'dart:convert';
import 'dart:io';

import 'package:flutter_test/flutter_test.dart';
import 'package:klp_guide/models/venues_response.dart';

void main() {
  test('VenuesResponse.fromJson parses envelope', () {
    final raw = File(
      'test/fixtures/venues_response_sample.json',
    ).readAsStringSync();
    final response = VenuesResponse.fromJson(
      jsonDecode(raw) as Map<String, dynamic>,
    );
    expect(response.total, 1);
    expect(response.venues, hasLength(1));
    expect(response.venues.first.name, 'SCHLANZE');
  });

  test('VenuesResponse.fromJson handles empty list', () {
    final response = VenuesResponse.fromJson({
      'venues': <dynamic>[],
      'total': 0,
    });
    expect(response.total, 0);
    expect(response.venues, isEmpty);
  });
}
