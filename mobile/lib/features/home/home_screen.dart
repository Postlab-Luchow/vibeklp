import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../models/venue.dart';
import '../../providers/venues_provider.dart';

class HomeScreen extends ConsumerWidget {
  const HomeScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final venuesAsync = ref.watch(venuesProvider);
    return Scaffold(
      appBar: AppBar(title: const Text('KLP-Guide')),
      body: venuesAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (err, _) => _ErrorView(error: err),
        data: (venues) =>
            venues.isEmpty ? const _EmptyView() : _VenueList(venues: venues),
      ),
    );
  }
}

class _VenueList extends StatelessWidget {
  const _VenueList({required this.venues});
  final List<Venue> venues;

  @override
  Widget build(BuildContext context) {
    return ListView.builder(
      itemCount: venues.length,
      itemBuilder: (context, index) {
        final v = venues[index];
        return ListTile(title: Text(v.name), subtitle: Text(v.address.city));
      },
    );
  }
}

class _ErrorView extends StatelessWidget {
  const _ErrorView({required this.error});
  final Object error;

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Text(
          'Fehler beim Laden der Daten:\n$error',
          textAlign: TextAlign.center,
        ),
      ),
    );
  }
}

class _EmptyView extends StatelessWidget {
  const _EmptyView();

  @override
  Widget build(BuildContext context) {
    return const Center(child: Text('Keine Veranstaltungsorte gefunden.'));
  }
}
