import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../models/venue.dart';
import '../repositories/venue_repo.dart';

class VenuesNotifier extends AsyncNotifier<List<Venue>> {
  @override
  Future<List<Venue>> build() {
    final repo = ref.watch(venueRepositoryProvider);
    return repo.fetchAll();
  }
}

final venuesProvider = AsyncNotifierProvider<VenuesNotifier, List<Venue>>(
  VenuesNotifier.new,
);
