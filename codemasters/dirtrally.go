package codemasters

// http://forums.codemasters.com/discussion/46726/d-box-and-udp-telemetry-information
// https://forums.codemasters.com/topic/16703-d-box-and-udp-telemetry-information/
//https://docs.google.com/spreadsheets/d/1UTgeE7vbnGIzDz-URRk2eBIPc_LR1vWcZklp7xD9N0Y/edit#gid=0

import (
	"encoding/binary"
	"math"
)

// DirtPacketSize is 264 bytes (66 * 32-bit words/fields/floats) [extradata=3]
const DirtPacketSize = 264

// DirtPacket is a bit shy of 264 bytes, so clearly missing some data.
type DirtPacket struct {
	Time           float32
	LapTime        float32
	LapDistance    float32
	TotalDistance  float32
	X              float32 // World space position
	Y              float32 // World space position
	Z              float32 // World space position
	Speed          float32
	Xv             float32 // Velocity in world space
	Yv             float32 // Velocity in world space
	Zv             float32 // Velocity in world space
	Xr             float32 // World space right direction
	Yr             float32 // World space right direction
	Zr             float32 // World space right direction
	Xd             float32 // World space forward direction
	Yd             float32 // World space forward direction
	Zd             float32 // World space forward direction
	Susp_pos_bl    float32
	Susp_pos_br    float32
	Susp_pos_fl    float32
	Susp_pos_fr    float32
	Susp_vel_bl    float32
	Susp_vel_br    float32
	Susp_vel_fl    float32
	Susp_vel_fr    float32
	Wheel_speed_bl float32
	Wheel_speed_br float32
	Wheel_speed_fl float32
	Wheel_speed_fr float32
	Throttle       float32
	Steer          float32
	Brake          float32
	Clutch         float32
	Gear           float32
	Gforce_lat     float32
	Gforce_lon     float32
	Lap            float32
	EngineRate     float32
	//############################################################# unknown start
	Sli_pro_native_support float32 // SLI Pro support
	Car_position           float32 // car race position
	Kers_level             float32 // kers energy left
	Kers_max_level         float32 // kers maximum energy
	Drs                    float32 // 0 = off, 1 = on
	Traction_control       float32 // 0 (off) - 2 (high)
	Anti_lock_brakes       float32 // 0 (off) - 1 (on)
	Fuel_in_tank           float32 // current fuel mass
	Fuel_capacity          float32 // fuel capacity
	In_pits                float32 // 0 = none, 1 = pitting, 2 = in pit area
	Sector                 float32 // 0 = sector1, 1 = sector2 2 = sector3
	Sector1_time           float32 // time of sector1 (or 0)
	Sector2_time           float32 // time of sector2 (or 0)
	//############################################################# unknown end
	Brakes_temp [4]float32 // brakes temperature (centigrade)
	//############################################################# unknown start
	Wheels_pressure [4]float32 // wheels pressure PSI
	Team_info       float32    // team ID
	//############################################################# unknown end
	Total_laps    float32 // total number of laps in this race
	Track_size    float32 // track size meters
	Last_lap_time float32 // last lap time
	Max_rpm       float32 // cars max RPM, at which point the rev limiter will kick in
	//Idle_rpm               float32    // cars idle RPM
	//Max_gears              float32    // maximum number of gears
	//SessionType            float32    // 0 = unknown, 1 = practice, 2 = qualifying, 3 = race
	//DrsAllowed             float32    // 0 = not allowed, 1 = allowed, -1 = invalid / unknown
	//Track_number           float32    // -1 for unknown, 0-21 for tracks
	//VehicleFIAFlags        float32    // -1 = invalid/unknown, 0 = none, 1 = green, 2 = blue, 3 = yellow, 4 = red
}

func (p *DirtPacket) Size() int {
	return DirtPacketSize
}

func (p *DirtPacket) GetGear() int {
	return int(p.Gear)
}

func (p *DirtPacket) GetRevLightPercent() int {
	return int((100 * p.EngineRate) / p.Max_rpm)
}

func (p *DirtPacket) GetSpeed() int {
	// TODO
	return 0
}

// Decode converts a little endian byte array into a DirtPacket.  Although this
// is fairly verbose, it is far far faster than using binary.Read() since it
// involves no allocations or reflection.
func (p *DirtPacket) Decode(b []byte) {
	_ = b[263] // bounds check hint to compiler; see golang.org/issue/14808
	p.Time = math.Float32frombits(binary.LittleEndian.Uint32(b[:4]))
	p.LapTime = math.Float32frombits(binary.LittleEndian.Uint32(b[4:8]))
	p.LapDistance = math.Float32frombits(binary.LittleEndian.Uint32(b[8:12]))
	p.TotalDistance = math.Float32frombits(binary.LittleEndian.Uint32(b[12:16]))
	p.X = math.Float32frombits(binary.LittleEndian.Uint32(b[16:20]))
	p.Y = math.Float32frombits(binary.LittleEndian.Uint32(b[20:24]))
	p.Z = math.Float32frombits(binary.LittleEndian.Uint32(b[24:28]))
	p.Speed = math.Float32frombits(binary.LittleEndian.Uint32(b[28:32]))
	p.Xv = math.Float32frombits(binary.LittleEndian.Uint32(b[32:36]))
	p.Yv = math.Float32frombits(binary.LittleEndian.Uint32(b[36:40]))
	p.Zv = math.Float32frombits(binary.LittleEndian.Uint32(b[40:44]))
	p.Xr = math.Float32frombits(binary.LittleEndian.Uint32(b[44:48]))
	p.Yr = math.Float32frombits(binary.LittleEndian.Uint32(b[48:52]))
	p.Zr = math.Float32frombits(binary.LittleEndian.Uint32(b[52:56]))
	p.Xd = math.Float32frombits(binary.LittleEndian.Uint32(b[56:60]))
	p.Yd = math.Float32frombits(binary.LittleEndian.Uint32(b[60:64]))
	p.Zd = math.Float32frombits(binary.LittleEndian.Uint32(b[64:68]))
	p.Susp_pos_bl = math.Float32frombits(binary.LittleEndian.Uint32(b[68:72]))
	p.Susp_pos_br = math.Float32frombits(binary.LittleEndian.Uint32(b[72:76]))
	p.Susp_pos_fl = math.Float32frombits(binary.LittleEndian.Uint32(b[76:80]))
	p.Susp_pos_fr = math.Float32frombits(binary.LittleEndian.Uint32(b[80:84]))
	p.Susp_vel_bl = math.Float32frombits(binary.LittleEndian.Uint32(b[84:88]))
	p.Susp_vel_br = math.Float32frombits(binary.LittleEndian.Uint32(b[88:92]))
	p.Susp_vel_fl = math.Float32frombits(binary.LittleEndian.Uint32(b[92:96]))
	p.Susp_vel_fr = math.Float32frombits(binary.LittleEndian.Uint32(b[96:100]))
	p.Wheel_speed_bl = math.Float32frombits(binary.LittleEndian.Uint32(b[100:104]))
	p.Wheel_speed_br = math.Float32frombits(binary.LittleEndian.Uint32(b[104:108]))
	p.Wheel_speed_fl = math.Float32frombits(binary.LittleEndian.Uint32(b[108:112]))
	p.Wheel_speed_fr = math.Float32frombits(binary.LittleEndian.Uint32(b[112:116]))
	p.Throttle = math.Float32frombits(binary.LittleEndian.Uint32(b[116:120]))
	p.Steer = math.Float32frombits(binary.LittleEndian.Uint32(b[120:124]))
	p.Brake = math.Float32frombits(binary.LittleEndian.Uint32(b[124:128]))
	p.Clutch = math.Float32frombits(binary.LittleEndian.Uint32(b[128:132]))
	p.Gear = math.Float32frombits(binary.LittleEndian.Uint32(b[132:136]))
	p.Gforce_lat = math.Float32frombits(binary.LittleEndian.Uint32(b[136:140]))
	p.Gforce_lon = math.Float32frombits(binary.LittleEndian.Uint32(b[140:144]))
	p.Lap = math.Float32frombits(binary.LittleEndian.Uint32(b[144:148]))
	p.EngineRate = math.Float32frombits(binary.LittleEndian.Uint32(b[148:152]))
	p.Sli_pro_native_support = math.Float32frombits(binary.LittleEndian.Uint32(b[152:156]))
	p.Car_position = math.Float32frombits(binary.LittleEndian.Uint32(b[156:160]))
	p.Kers_level = math.Float32frombits(binary.LittleEndian.Uint32(b[160:164]))
	p.Kers_max_level = math.Float32frombits(binary.LittleEndian.Uint32(b[164:168]))
	p.Drs = math.Float32frombits(binary.LittleEndian.Uint32(b[168:172]))
	p.Traction_control = math.Float32frombits(binary.LittleEndian.Uint32(b[172:176]))
	p.Anti_lock_brakes = math.Float32frombits(binary.LittleEndian.Uint32(b[176:180]))
	p.Fuel_in_tank = math.Float32frombits(binary.LittleEndian.Uint32(b[180:184]))
	p.Fuel_capacity = math.Float32frombits(binary.LittleEndian.Uint32(b[184:188]))
	p.In_pits = math.Float32frombits(binary.LittleEndian.Uint32(b[188:192]))
	p.Sector = math.Float32frombits(binary.LittleEndian.Uint32(b[192:196]))
	p.Sector1_time = math.Float32frombits(binary.LittleEndian.Uint32(b[196:200]))
	p.Sector2_time = math.Float32frombits(binary.LittleEndian.Uint32(b[200:204]))
	p.Brakes_temp[0] = math.Float32frombits(binary.LittleEndian.Uint32(b[204:208]))
	p.Brakes_temp[1] = math.Float32frombits(binary.LittleEndian.Uint32(b[208:212]))
	p.Brakes_temp[2] = math.Float32frombits(binary.LittleEndian.Uint32(b[212:216]))
	p.Brakes_temp[3] = math.Float32frombits(binary.LittleEndian.Uint32(b[216:220]))
	p.Wheels_pressure[0] = math.Float32frombits(binary.LittleEndian.Uint32(b[220:224]))
	p.Wheels_pressure[1] = math.Float32frombits(binary.LittleEndian.Uint32(b[224:228]))
	p.Wheels_pressure[2] = math.Float32frombits(binary.LittleEndian.Uint32(b[228:232]))
	p.Wheels_pressure[3] = math.Float32frombits(binary.LittleEndian.Uint32(b[232:236]))
	p.Team_info = math.Float32frombits(binary.LittleEndian.Uint32(b[236:240]))
	p.Total_laps = math.Float32frombits(binary.LittleEndian.Uint32(b[240:244]))
	p.Track_size = math.Float32frombits(binary.LittleEndian.Uint32(b[244:248]))
	p.Last_lap_time = math.Float32frombits(binary.LittleEndian.Uint32(b[248:252]))
	p.Max_rpm = math.Float32frombits(binary.LittleEndian.Uint32(b[252:256]))
}
