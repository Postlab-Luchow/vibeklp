import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:klp_guide/models/venue.dart';
import 'package:klp_guide/providers/venues_provider.dart';
import 'package:klp_guide/repositories/venue_repo.dart';

class MockVenueRepository extends Mock implements VenueRepository {}

final _sampleVenues = [
  const Venue(
    id: 'v1',
    name: 'Alpha',
    address: Address(city: 'Lüchow'),
    coordinates: Coordinates(lat: 53.0, lng: 11.1),
  ),
  const Venue(
    id: 'v2',
    name: 'Beta',
    address: Address(city: 'Dannenberg'),
    coordinates: Coordinates(lat: 53.1, lng: 11.2),
  ),
];

void main() {
  test('venuesProvider exposes data from the repository', () async {
    final repo = MockVenueRepository();
    when(repo.fetchAll).thenAnswer((_) async => _sampleVenues);

    final container = ProviderContainer(
      overrides: [venueRepositoryProvider.overrideWithValue(repo)],
    );
    addTearDown(container.dispose);

    final result = await container.read(venuesProvider.future);
    expect(result, hasLength(2));
    expect(result.first.name, 'Alpha');
  });

  test('venuesProvider surfaces repository errors as AsyncError', () async {
    final repo = MockVenueRepository();
    // Throw asynchronously to model a real (network) failure: build() returns a
    // pending future, so the initial state is AsyncLoading before it settles to
    // AsyncError. A synchronous throw would skip the loading state entirely.
    when(repo.fetchAll).thenAnswer((_) async => throw StateError('boom'));

    final container = ProviderContainer(
      overrides: [venueRepositoryProvider.overrideWithValue(repo)],
    );
    addTearDown(container.dispose);

    final result = container.read(venuesProvider);
    expect(result.isLoading, isTrue);

    await Future<void>.delayed(Duration.zero);

    final settled = container.read(venuesProvider);
    expect(settled.hasError, isTrue);
    expect(settled.error, isA<StateError>());
  });
}
