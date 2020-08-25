package bit

type BitInt uint64

//文档中如果出现 形如 2#1010101 的值， 则表示2进制的1010101

// 此模块提供按位操作的功能，用于操作一个uint64
// pos 以0为基
//                                                                               pos第0位
//                                                                               ↓
//|0000 0000|0000 0000|0000 0000|0000 0000|0000 0000|0000 0000|0000 0000|0000 0000|
//                                                                              ↑
//                                                                              pos第1位
// pos 以0为基
// 将bitint的第pos位设置成1
func (bitInt *BitInt) Set(pos uint64) *BitInt {
	if pos > 63 {
		return bitInt
	}
	*bitInt = *bitInt | (1 << pos)
	return bitInt
}

// 将bitint的第pos位设置成0
func (bitInt *BitInt) Unset(pos uint64) *BitInt {
	if pos > 63 {
		return bitInt
	}
	*bitInt = *bitInt & (^(1 << pos))
	return bitInt
}

// 获取第pos位的值，返回值为0 或 1
func (bitInt BitInt) Get(pos uint64) uint64 {
	if pos > 63 {
		return 0
	}
	return uint64((bitInt >> pos) & 1)
}

// 判断第pos位是不是1 ,是1则返回true
func (bitInt BitInt) IsSet(pos uint64) bool {
	if pos > 63 {
		return false
	}
	return ((bitInt >> pos) & 1) == 1
}

//                                pos第2位
//                                ↓
// 2#00000000000000000000000000000000
// after SetValue(2,3,7) //将从pos=2的位置开始 往左数3数的区域设置成value对应的值
// 2#00000000000000000000000000011100
// GetValue(2,3)=2#111=7
func (bitInt *BitInt) SetValue(pos int, bitCount int, value uint64) *BitInt {
	if pos+bitCount-1 > 63 {
		return bitInt
	}
	valueBit := BitInt(value)
	for i := 0; i < bitCount; i++ {
		if valueBit.Get(uint64(i)) == 1 {
			bitInt.Set(uint64(pos + i))
		} else {
			bitInt.Unset(uint64(pos + i))
		}
	}
	return bitInt
}
func (bitInt BitInt) GetValue(pos int, bitCount int) uint64 {
	if pos+bitCount-1 > 63 {
		return 0
	}
	return uint64((bitInt >> uint64(pos)) & ((1 << uint64(bitCount)) - 1))
}

// 64位，则为8个字节，此函数用于获取第n个字节对应的字节值
// | 64bit | 0000 0000 | 0000 0000 | 0000 0000 | 0000 0000 | 0000 0000 | 0000 0000 | 0000 0000 | 0000 0000 |
// |  idx  |         7 |         6 |         5 |         4 |         3 |         2 |         1 |         0 |
// byte可取值范围如上图 [0,7]
func (bitInt BitInt) GetByte(idx int) uint8 {
	if idx < 0 || idx > 7 {
		return 0
	}
	// 	return uint8(bitInt.GetValue(idx*8, 8))
	return uint8((uint64(bitInt) >> (uint64(idx) * 8)) & 0x000000ff)
}
func (bitInt *BitInt) SetByte(idx int, value uint8) bool {
	if idx < 0 || idx > 7 {
		return false
	}
	bitInt.SetValue(idx*8, 8, uint64(value))
	return true
}

// 64位，则为4*16个，此函数用于获取第idx个uint16对应的字节值
// | 64bit | 0000 0000  0000 0000 | 0000 0000  0000 0000 | 0000 0000  0000 0000 | 0000 0000  0000 0000 |
// | idx   |                    3 |                    2 |                    1 |                    0 |
// u16可取值范围如上图 [0,3]
func (bitInt BitInt) GetUint16(idx int) uint16 {
	if idx < 0 || idx > 3 {
		return 0
	}
	// 	return uint16(bitInt.GetValue(idx*16, 16))
	return uint16((uint64(bitInt) >> (uint64(idx) * 16)) & 0x0000ffff)
}
func (bitInt *BitInt) SetUint16(idx int, value uint16) bool {
	if idx < 0 || idx > 3 {
		return false
	}
	bitInt.SetValue(idx*16, 16, uint64(value))
	return true
}

// 64位，则为2*32个，此函数用于获取第idx个uint32对应的字节值
// | 64bit | 0000 0000  0000 0000  0000 0000  0000 0000 | 0000 0000  0000 0000  0000 0000  0000 0000 |
// | idx   |                                          1 |                                          0 |
// idx可取值范围如上图 [0,1]
func (bitInt BitInt) GetUint32(idx int) uint32 {
	if idx < 0 || idx > 1 {
		return 0
	}
	return uint32((uint64(bitInt) >> (uint64(idx) * 32)) & 0xffffffff)
	// return uint32(bitInt.GetValue(idx*32, 32))
}
func (bitInt *BitInt) SetUint32(idx int, value uint32) bool {
	if idx < 0 || idx > 4 {
		return false
	}
	bitInt.SetValue(idx*32, 32, uint64(value))
	return true
}

func (bitInt BitInt) Len() int {
	return 64
}
func (bitInt BitInt) IsZero() bool {
	return bitInt == 0
}

func (bitInt *BitInt) GetFlagCount(mostPos ...uint64) (cnt uint64) { //
	// see TestBitIntGetFlagCount
	// 返回共有多少位是1,可选参数mostPos ,指定不必检测所有的63位pos ,只需要检测指定个数 的位置即可
	// 如mostPos 为10 ， 则只检测0~9位的数字
	var total uint64 = 64
	var pos uint64 = 0
	if len(mostPos) > 0 && mostPos[0] < total {
		total = mostPos[0]
	}
	for pos = 0; pos < total; pos++ {
		if bitInt.IsSet(pos) {
			cnt++
		}
	}
	return
}

func (bitInt BitInt) ToUint64() uint64 {
	return uint64(bitInt)
}
