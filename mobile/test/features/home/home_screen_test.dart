import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:mocktail/mocktail.dart';

import 'package:klp_guide/features/home/home_screen.dart';
import 'package:klp_guide/models/venue.dart';
import 'package:klp_guide/repositories/venue_repo.dart';

class MockVenueRepository extends Mock implements VenueRepository {}

Future<void> _pumpHome(WidgetTester tester, VenueRepository repo) {
  return tester.pumpWidget(
    ProviderScope(
      overrides: [venueRepositoryProvider.overrideWithValue(repo)],
      child: const MaterialApp(home: HomeScreen()),
    ),
  );
}

void main() {
  testWidgets('shows progress indicator while loading', (tester) async {
    final repo = MockVenueRepository();
    when(repo.fetchAll).thenAnswer(
      (_) => Future<List<Venue>>.delayed(const Duration(milliseconds: 200), () => []),
    );

    await _pumpHome(tester, repo);
    expect(find.byType(CircularProgressIndicator), findsOneWidget);

    // Drain the pending future so the test exits cleanly
    await tester.pump(const Duration(milliseconds: 250));
  });

  testWidgets('shows venue names when data loads', (tester) async {
    final repo = MockVenueRepository();
    when(repo.fetchAll).thenAnswer(
      (_) async => const [
        Venue(
          id: 'v1',
          name: 'SCHLANZE',
          address: Address(city: 'Waddeweitz'),
          coordinates: Coordinates(lat: 53.0, lng: 11.0),
        ),
        Venue(
          id: 'v2',
          name: 'LÜBELNER MÜHLE',
          address: Address(city: 'Küsten'),
          coordinates: Coordinates(lat: 53.0, lng: 11.1),
        ),
      ],
    );

    await _pumpHome(tester, repo);
    await tester.pumpAndSettle();

    expect(find.text('SCHLANZE'), findsOneWidget);
    expect(find.text('LÜBELNER MÜHLE'), findsOneWidget);
    expect(find.text('Waddeweitz'), findsOneWidget);
  });

  testWidgets('shows error message when fetch fails', (tester) async {
    final repo = MockVenueRepository();
    when(repo.fetchAll).thenAnswer((_) async => throw StateError('network down'));

    await _pumpHome(tester, repo);
    await tester.pumpAndSettle();

    expect(find.textContaining('Fehler'), findsOneWidget);
  });

  testWidgets('shows empty-state message when list is empty', (tester) async {
    final repo = MockVenueRepository();
    when(repo.fetchAll).thenAnswer((_) async => <Venue>[]);

    await _pumpHome(tester, repo);
    await tester.pumpAndSettle();

    expect(find.textContaining('Keine Veranstaltungsorte'), findsOneWidget);
  });
}
