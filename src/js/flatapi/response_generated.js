// automatically generated by the FlatBuffers compiler, do not modify

/**
 * @const
*/
var flatapi = flatapi || {};

/**
 * @enum
 */
flatapi.AnyResult = {
  NONE: 0,
  QueryResult: 1
};

/**
 * @constructor
 */
flatapi.Result = function() {
  /**
   * @type {flatbuffers.ByteBuffer}
   */
  this.bb = null;

  /**
   * @type {number}
   */
  this.bb_pos = 0;
};

/**
 * @param {number} i
 * @param {flatbuffers.ByteBuffer} bb
 * @returns {flatapi.Result}
 */
flatapi.Result.prototype.__init = function(i, bb) {
  this.bb_pos = i;
  this.bb = bb;
  return this;
};

/**
 * @param {flatbuffers.ByteBuffer} bb
 * @param {flatapi.Result=} obj
 * @returns {flatapi.Result}
 */
flatapi.Result.getRootAsResult = function(bb, obj) {
  return (obj || new flatapi.Result).__init(bb.readInt32(bb.position()) + bb.position(), bb);
};

/**
 * @returns {flatapi.AnyResult}
 */
flatapi.Result.prototype.resultType = function() {
  var offset = this.bb.__offset(this.bb_pos, 4);
  return offset ? /** @type {flatapi.AnyResult} */ (this.bb.readUint8(this.bb_pos + offset)) : flatapi.AnyResult.NONE;
};

/**
 * @param {flatbuffers.Table} obj
 * @returns {?flatbuffers.Table}
 */
flatapi.Result.prototype.result = function(obj) {
  var offset = this.bb.__offset(this.bb_pos, 6);
  return offset ? this.bb.__union(obj, this.bb_pos + offset) : null;
};

/**
 * @param {flatbuffers.Builder} builder
 */
flatapi.Result.startResult = function(builder) {
  builder.startObject(2);
};

/**
 * @param {flatbuffers.Builder} builder
 * @param {flatapi.AnyResult} resultType
 */
flatapi.Result.addResultType = function(builder, resultType) {
  builder.addFieldInt8(0, resultType, flatapi.AnyResult.NONE);
};

/**
 * @param {flatbuffers.Builder} builder
 * @param {flatbuffers.Offset} resultOffset
 */
flatapi.Result.addResult = function(builder, resultOffset) {
  builder.addFieldOffset(1, resultOffset, 0);
};

/**
 * @param {flatbuffers.Builder} builder
 * @returns {flatbuffers.Offset}
 */
flatapi.Result.endResult = function(builder) {
  var offset = builder.endObject();
  return offset;
};

/**
 * @constructor
 */
flatapi.Response = function() {
  /**
   * @type {flatbuffers.ByteBuffer}
   */
  this.bb = null;

  /**
   * @type {number}
   */
  this.bb_pos = 0;
};

/**
 * @param {number} i
 * @param {flatbuffers.ByteBuffer} bb
 * @returns {flatapi.Response}
 */
flatapi.Response.prototype.__init = function(i, bb) {
  this.bb_pos = i;
  this.bb = bb;
  return this;
};

/**
 * @param {flatbuffers.ByteBuffer} bb
 * @param {flatapi.Response=} obj
 * @returns {flatapi.Response}
 */
flatapi.Response.getRootAsResponse = function(bb, obj) {
  return (obj || new flatapi.Response).__init(bb.readInt32(bb.position()) + bb.position(), bb);
};

/**
 * @param {flatbuffers.Encoding=} optionalEncoding
 * @returns {string|Uint8Array}
 */
flatapi.Response.prototype.id = function(optionalEncoding) {
  var offset = this.bb.__offset(this.bb_pos, 4);
  return offset ? this.bb.__string(this.bb_pos + offset, optionalEncoding) : null;
};

/**
 * @param {number} index
 * @param {flatapi.Result=} obj
 * @returns {flatapi.Result}
 */
flatapi.Response.prototype.result = function(index, obj) {
  var offset = this.bb.__offset(this.bb_pos, 6);
  return offset ? (obj || new flatapi.Result).__init(this.bb.__indirect(this.bb.__vector(this.bb_pos + offset) + index * 4), this.bb) : null;
};

/**
 * @returns {number}
 */
flatapi.Response.prototype.resultLength = function() {
  var offset = this.bb.__offset(this.bb_pos, 6);
  return offset ? this.bb.__vector_len(this.bb_pos + offset) : 0;
};

/**
 * @param {flatbuffers.Builder} builder
 */
flatapi.Response.startResponse = function(builder) {
  builder.startObject(2);
};

/**
 * @param {flatbuffers.Builder} builder
 * @param {flatbuffers.Offset} idOffset
 */
flatapi.Response.addId = function(builder, idOffset) {
  builder.addFieldOffset(0, idOffset, 0);
};

/**
 * @param {flatbuffers.Builder} builder
 * @param {flatbuffers.Offset} resultOffset
 */
flatapi.Response.addResult = function(builder, resultOffset) {
  builder.addFieldOffset(1, resultOffset, 0);
};

/**
 * @param {flatbuffers.Builder} builder
 * @param {Array.<flatbuffers.Offset>} data
 * @returns {flatbuffers.Offset}
 */
flatapi.Response.createResultVector = function(builder, data) {
  builder.startVector(4, data.length, 4);
  for (var i = data.length - 1; i >= 0; i--) {
    builder.addOffset(data[i]);
  }
  return builder.endVector();
};

/**
 * @param {flatbuffers.Builder} builder
 * @param {number} numElems
 */
flatapi.Response.startResultVector = function(builder, numElems) {
  builder.startVector(4, numElems, 4);
};

/**
 * @param {flatbuffers.Builder} builder
 * @returns {flatbuffers.Offset}
 */
flatapi.Response.endResponse = function(builder) {
  var offset = builder.endObject();
  return offset;
};

/**
 * @param {flatbuffers.Builder} builder
 * @param {flatbuffers.Offset} offset
 */
flatapi.Response.finishResponseBuffer = function(builder, offset) {
  builder.finish(offset);
};

// Exports for Node.js and RequireJS
this.flatapi = flatapi;
