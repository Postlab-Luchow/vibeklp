// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'venues_response.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

T _$identity<T>(T value) => value;

final _privateConstructorUsedError = UnsupportedError(
  'It seems like you constructed your class using `MyClass._()`. This constructor is only meant to be used by freezed and you are not supposed to need it nor use it.\nPlease check the documentation here for more information: https://github.com/rrousselGit/freezed#adding-getters-and-methods-to-our-models',
);

VenuesResponse _$VenuesResponseFromJson(Map<String, dynamic> json) {
  return _VenuesResponse.fromJson(json);
}

/// @nodoc
mixin _$VenuesResponse {
  List<Venue> get venues => throw _privateConstructorUsedError;
  int get total => throw _privateConstructorUsedError;

  /// Serializes this VenuesResponse to a JSON map.
  Map<String, dynamic> toJson() => throw _privateConstructorUsedError;

  /// Create a copy of VenuesResponse
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  $VenuesResponseCopyWith<VenuesResponse> get copyWith =>
      throw _privateConstructorUsedError;
}

/// @nodoc
abstract class $VenuesResponseCopyWith<$Res> {
  factory $VenuesResponseCopyWith(
    VenuesResponse value,
    $Res Function(VenuesResponse) then,
  ) = _$VenuesResponseCopyWithImpl<$Res, VenuesResponse>;
  @useResult
  $Res call({List<Venue> venues, int total});
}

/// @nodoc
class _$VenuesResponseCopyWithImpl<$Res, $Val extends VenuesResponse>
    implements $VenuesResponseCopyWith<$Res> {
  _$VenuesResponseCopyWithImpl(this._value, this._then);

  // ignore: unused_field
  final $Val _value;
  // ignore: unused_field
  final $Res Function($Val) _then;

  /// Create a copy of VenuesResponse
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({Object? venues = null, Object? total = null}) {
    return _then(
      _value.copyWith(
            venues: null == venues
                ? _value.venues
                : venues // ignore: cast_nullable_to_non_nullable
                      as List<Venue>,
            total: null == total
                ? _value.total
                : total // ignore: cast_nullable_to_non_nullable
                      as int,
          )
          as $Val,
    );
  }
}

/// @nodoc
abstract class _$$VenuesResponseImplCopyWith<$Res>
    implements $VenuesResponseCopyWith<$Res> {
  factory _$$VenuesResponseImplCopyWith(
    _$VenuesResponseImpl value,
    $Res Function(_$VenuesResponseImpl) then,
  ) = __$$VenuesResponseImplCopyWithImpl<$Res>;
  @override
  @useResult
  $Res call({List<Venue> venues, int total});
}

/// @nodoc
class __$$VenuesResponseImplCopyWithImpl<$Res>
    extends _$VenuesResponseCopyWithImpl<$Res, _$VenuesResponseImpl>
    implements _$$VenuesResponseImplCopyWith<$Res> {
  __$$VenuesResponseImplCopyWithImpl(
    _$VenuesResponseImpl _value,
    $Res Function(_$VenuesResponseImpl) _then,
  ) : super(_value, _then);

  /// Create a copy of VenuesResponse
  /// with the given fields replaced by the non-null parameter values.
  @pragma('vm:prefer-inline')
  @override
  $Res call({Object? venues = null, Object? total = null}) {
    return _then(
      _$VenuesResponseImpl(
        venues: null == venues
            ? _value._venues
            : venues // ignore: cast_nullable_to_non_nullable
                  as List<Venue>,
        total: null == total
            ? _value.total
            : total // ignore: cast_nullable_to_non_nullable
                  as int,
      ),
    );
  }
}

/// @nodoc
@JsonSerializable()
class _$VenuesResponseImpl implements _VenuesResponse {
  const _$VenuesResponseImpl({
    required final List<Venue> venues,
    required this.total,
  }) : _venues = venues;

  factory _$VenuesResponseImpl.fromJson(Map<String, dynamic> json) =>
      _$$VenuesResponseImplFromJson(json);

  final List<Venue> _venues;
  @override
  List<Venue> get venues {
    if (_venues is EqualUnmodifiableListView) return _venues;
    // ignore: implicit_dynamic_type
    return EqualUnmodifiableListView(_venues);
  }

  @override
  final int total;

  @override
  String toString() {
    return 'VenuesResponse(venues: $venues, total: $total)';
  }

  @override
  bool operator ==(Object other) {
    return identical(this, other) ||
        (other.runtimeType == runtimeType &&
            other is _$VenuesResponseImpl &&
            const DeepCollectionEquality().equals(other._venues, _venues) &&
            (identical(other.total, total) || other.total == total));
  }

  @JsonKey(includeFromJson: false, includeToJson: false)
  @override
  int get hashCode => Object.hash(
    runtimeType,
    const DeepCollectionEquality().hash(_venues),
    total,
  );

  /// Create a copy of VenuesResponse
  /// with the given fields replaced by the non-null parameter values.
  @JsonKey(includeFromJson: false, includeToJson: false)
  @override
  @pragma('vm:prefer-inline')
  _$$VenuesResponseImplCopyWith<_$VenuesResponseImpl> get copyWith =>
      __$$VenuesResponseImplCopyWithImpl<_$VenuesResponseImpl>(
        this,
        _$identity,
      );

  @override
  Map<String, dynamic> toJson() {
    return _$$VenuesResponseImplToJson(this);
  }
}

abstract class _VenuesResponse implements VenuesResponse {
  const factory _VenuesResponse({
    required final List<Venue> venues,
    required final int total,
  }) = _$VenuesResponseImpl;

  factory _VenuesResponse.fromJson(Map<String, dynamic> json) =
      _$VenuesResponseImpl.fromJson;

  @override
  List<Venue> get venues;
  @override
  int get total;

  /// Create a copy of VenuesResponse
  /// with the given fields replaced by the non-null parameter values.
  @override
  @JsonKey(includeFromJson: false, includeToJson: false)
  _$$VenuesResponseImplCopyWith<_$VenuesResponseImpl> get copyWith =>
      throw _privateConstructorUsedError;
}
