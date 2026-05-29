import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../core/api_client.dart';
import '../models/venue.dart';
import '../models/venues_response.dart';

class VenueRepository {
  VenueRepository(this._dio);

  final Dio _dio;

  Future<List<Venue>> fetchAll() async {
    final response = await _dio.get<Map<String, dynamic>>('/api/venues');
    final body = response.data;
    if (body == null) {
      throw StateError('GET /api/venues returned empty body');
    }
    return VenuesResponse.fromJson(body).venues;
  }
}

final venueRepositoryProvider = Provider<VenueRepository>(
  (ref) => VenueRepository(ref.watch(apiClientProvider)),
);
