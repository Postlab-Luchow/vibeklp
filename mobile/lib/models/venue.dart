import 'package:freezed_annotation/freezed_annotation.dart';

part 'venue.freezed.dart';
part 'venue.g.dart';

@freezed
class Venue with _$Venue {
  const factory Venue({
    required String id,
    required String name,
    @Default('') String description,
    required Address address,
    required Coordinates coordinates,
    @Default(Contact()) Contact contact,
    @Default(<String>[]) List<String> amenities,
    @Default(<VenueCategory>[]) List<VenueCategory> categories,
    @Default('') String bikeRoute,
    @Default('') String source,
    @Default(<String>[]) List<String> eventIds,
    @Default(<String>[]) List<String> exhibitionIds,
    @Default(0) int eventCount,
    @Default(0) int exhibitionCount,
  }) = _Venue;

  factory Venue.fromJson(Map<String, dynamic> json) => _$VenueFromJson(json);
}

@freezed
class Address with _$Address {
  const factory Address({
    @Default('') String street,
    @Default('') String postalCode,
    @Default('') String city,
  }) = _Address;

  factory Address.fromJson(Map<String, dynamic> json) =>
      _$AddressFromJson(json);
}

@freezed
class Coordinates with _$Coordinates {
  const factory Coordinates({required double lat, required double lng}) =
      _Coordinates;

  factory Coordinates.fromJson(Map<String, dynamic> json) =>
      _$CoordinatesFromJson(json);
}

@freezed
class Contact with _$Contact {
  const factory Contact({
    @Default('') String phone,
    @Default('') String email,
    @Default('') String website,
  }) = _Contact;

  factory Contact.fromJson(Map<String, dynamic> json) =>
      _$ContactFromJson(json);
}

@freezed
class VenueCategory with _$VenueCategory {
  const factory VenueCategory({
    required String name,
    @Default(<String>[]) List<String> dates,
  }) = _VenueCategory;

  factory VenueCategory.fromJson(Map<String, dynamic> json) =>
      _$VenueCategoryFromJson(json);
}
