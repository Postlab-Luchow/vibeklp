import 'package:freezed_annotation/freezed_annotation.dart';

import 'venue.dart';

part 'venues_response.freezed.dart';
part 'venues_response.g.dart';

@freezed
class VenuesResponse with _$VenuesResponse {
  const factory VenuesResponse({
    required List<Venue> venues,
    required int total,
  }) = _VenuesResponse;

  factory VenuesResponse.fromJson(Map<String, dynamic> json) =>
      _$VenuesResponseFromJson(json);
}
